// Copyright 2015 The appc Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lib

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/aci"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/discovery"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema/types"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/rkt/store"

	"github.com/appc/acbuild/util"
)

const (
	acbuildStorePath = "~/.acbuild"
)

var pathlist = []string{"/usr/local/sbin", "/usr/local/bin", "/usr/sbin",
	"/usr/bin", "/sbin", "/bin"}

// Run will execute the given command in the ACI being built. a.CurrentACIPath
// is where the untarred ACI is stored, a.DepStoreTarPath is the directory to
// download dependencies into, a.DepStoreExpandedPath is where the dependencies
// are expanded into, a.OverlayWorkPath is the work directory used by
// overlayfs, and insecure signifies whether downloaded images should be
// fetched over http or https.
func (a *ACBuild) Run(cmd []string, insecure bool) (err error) {
	if err = a.lock(); err != nil {
		return err
	}
	defer func() {
		if err1 := a.unlock(); err == nil {
			err = err1
		}
	}()

	if os.Geteuid() != 0 {
		return fmt.Errorf("the run subcommand must be run as root")
	}

	err = util.RmAndMkdir(a.OverlayTargetPath)
	if err != nil {
		return err
	}
	defer os.RemoveAll(a.OverlayTargetPath)
	err = util.RmAndMkdir(a.OverlayWorkPath)
	if err != nil {
		return err
	}
	defer os.RemoveAll(a.OverlayWorkPath)

	man, err := util.GetManifest(a.CurrentACIPath)
	if err != nil {
		return err
	}

	if len(man.Dependencies) != 0 {
		if !supportsOverlay() {
			err := exec.Command("modprobe", "overlay").Run()
			if err != nil {
				return err
			}
			if !supportsOverlay() {
				return fmt.Errorf(
					"overlayfs support required for using run with dependencies")
			}
		}
	}

	deps, err := a.renderACI(insecure, a.Debug)
	if err != nil {
		return err
	}

	var nspawnpath string
	if deps == nil {
		nspawnpath = path.Join(a.CurrentACIPath, aci.RootfsDir)
	} else {
		for i, dep := range deps {
			deps[i] = path.Join(a.DepStoreExpandedPath, dep, aci.RootfsDir)
		}
		options := "lowerdir=" + strings.Join(deps, ":") +
			",upperdir=" + path.Join(a.CurrentACIPath, aci.RootfsDir) +
			",workdir=" + a.OverlayWorkPath
		err := syscall.Mount("overlay", a.OverlayTargetPath, "overlay", 0, options)
		if err != nil {
			return err
		}

		defer func() {
			err1 := syscall.Unmount(a.OverlayTargetPath, 0)
			if err == nil {
				err = err1
			}
		}()

		nspawnpath = a.OverlayTargetPath
	}
	nspawncmd := []string{"systemd-nspawn", "-D", nspawnpath}

	version, err := getSystemdVersion()
	if err != nil {
		return err
	}
	if version >= 209 {
		nspawncmd = append(nspawncmd, "--quiet", "--register=no")
	}

	if man.App != nil {
		for _, evar := range man.App.Environment {
			nspawncmd = append(nspawncmd, "--setenv", evar.Name+"="+evar.Value)
		}
	}

	err = a.mirrorLocalZoneInfo()
	if err != nil {
		return err
	}

	if len(cmd) == 0 {
		return fmt.Errorf("command to run not set")
	}
	abscmd, err := findCmdInPath(pathlist, cmd[0], nspawnpath)
	if err != nil {
		return err
	}
	nspawncmd = append(nspawncmd, abscmd)
	nspawncmd = append(nspawncmd, cmd[1:]...)

	execCmd := exec.Command(nspawncmd[0], nspawncmd[1:]...)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	execCmd.Env = []string{"SYSTEMD_LOG_LEVEL=err"}

	err = execCmd.Run()
	if err != nil {
		if err == exec.ErrNotFound {
			return fmt.Errorf("systemd-nspawn is required but not found")
		}
		return err
	}

	return nil
}

// stolen from github.com/coreos/rkt/common/common.go
// supportsOverlay returns whether the system supports overlay filesystem
func supportsOverlay() bool {
	f, err := os.Open("/proc/filesystems")
	if err != nil {
		fmt.Println("error opening /proc/filesystems")
		return false
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		if s.Text() == "nodev\toverlay" {
			return true
		}
	}
	return false
}

func findCmdInPath(pathlist []string, cmd, prefix string) (string, error) {
	if path.IsAbs(cmd) {
		return cmd, nil
	}

	for _, p := range pathlist {
		_, err := os.Lstat(path.Join(prefix, p, cmd))
		switch {
		case os.IsNotExist(err):
			continue
		case err != nil:
			return "", err
		}
		return path.Join(p, cmd), nil
	}
	return "", fmt.Errorf("%s not found in any of: %v", cmd, pathlist)
}

func (a *ACBuild) mirrorLocalZoneInfo() error {
	zif, err := filepath.EvalSymlinks("/etc/localtime")
	if err != nil {
		return err
	}

	src, err := os.Open(zif)
	if err != nil {
		return err
	}
	defer src.Close()

	destp := filepath.Join(a.CurrentACIPath, aci.RootfsDir, zif)

	if err = os.MkdirAll(filepath.Dir(destp), 0755); err != nil {
		return err
	}

	dest, err := os.OpenFile(destp, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, src)
	if err != nil {
		return err
	}

	return nil
}

func getSystemdVersion() (int, error) {
	_, err := exec.LookPath("systemctl")
	if err == exec.ErrNotFound {
		return 0, fmt.Errorf("system does not have systemd")
	}

	blob, err := exec.Command("systemctl", "--version").Output()
	if err != nil {
		return 0, err
	}
	for _, line := range strings.Split(string(blob), "\n") {
		if strings.HasPrefix(line, "systemd ") {
			var version int
			_, err := fmt.Sscanf(line, "systemd %d", &version)
			if err != nil {
				return 0, err
			}
			return version, nil
		}
	}
	return 0, fmt.Errorf("error parsing output from `systemctl --version`")
}

func (a *ACBuild) renderACI(insecure, debug bool) ([]string, error) {
	man, err := util.GetManifest(a.CurrentACIPath)
	if err != nil {
		return nil, err
	}

	if len(man.Dependencies) == 0 {
		return nil, nil
	}

	s, err := store.NewStore(acbuildStorePath)
	if err != nil {
		return nil, err
	}

	var depList []string
	for _, dep := range man.Dependencies {
		err := a.fetchACI(s, dep, true)
		if err != nil {
			return nil, err
		}
		id, _, err := s.RenderTreeStore(key, false)
		if err != nil {
			return nil, err
		}
		depList = append(depList, s.GetTreeStoreRootFS(id))
	}
	return depList, nil
}

func (a *ACBuild) fetchACI(s *store.Store, aci types.Dependency, fetchDeps bool) error {
	endpoint, err := a.discoverEndpoint(aci.ImageName, aci.Labels)
	if err != nil {
		return err
	}

	remote, found, err := s.GetRemote(endpoint.ACI)
	if err != nil {
		return err
	}
	if !found {
		remote = store.NewRemote(endpoint.ACI, endpoint.ASC)
	}

	acifile, err := s.aciFile()
	if err != nil {
		return err
	}
	defer acifile.Close()
	defer os.Remove(aciFile.Name())

	remote.DownloadTime = time.Now()

	etag, err = a.download(remote.ACIURL, acifile, string(aci.ImageName))
	if err != nil {
		return err
	}

	remote.ETag = etag

	//TODO: download ASC, verify the ACI with it

	_, err = acifile.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}

	latest := false
	if v, ok := labels.Get("version"); !ok || v == "latest" {
		latest = true
	}

	key, err := s.WriteACI(acifile, latest)
	if err != nil {
		return err
	}

	remote.BlobKey = key

	err = s.WriteRemote(remote)
	if err != nil {
		return err
	}

	if !fetchDeps {
		return nil
	}

	man, err := s.GetImageManifest(key)
	if err != nil {
		return err
	}

	for _, dep := range man.Dependencies {
		err := a.fetchACIWithDeps(s, dep)
	}

	return nil
}

func (a *ACBuild) discoverEndpoint(imgName types.ACIdentifier, labels types.Labels) (*discovery.ACIEndpoint, error) {
	app, err := discovery.NewApp(string(imageName), labels.ToMap())
	if err != nil {
		return nil, err
	}
	if _, ok := app.Labels["arch"]; !ok {
		app.Labels["arch"] = runtime.GOARCH
	}
	if _, ok := app.Labels["os"]; !ok {
		app.Labels["os"] = runtime.GOOS
	}

	eps, attempts, err := discovery.DiscoverEndpoints(*app, a.Insecure)
	if err != nil {
		return nil, err
	}
	if len(eps.ACIEndpoints) == 0 {
		return nil, fmt.Errorf("no endpoints discovered to download %s", imageName)
	}

	return &eps.ACIEndpoints[0], nil
}

func (a *ACBuild) download(url string, writer io.Writer, label string) (string, error) {
	//TODO: auth
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	transport := http.DefaultTransport
	if r.Insecure {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	client := &http.Client{Transport: transport}
	//f.setHTTPHeaders(req, etag)

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return fmt.Errorf("too many redirects")
		}
		//f.setHTTPHeaders(req, etag)
		return nil
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		break
	default:
		return "", fmt.Errorf("bad HTTP status code: %d", res.StatusCode)
	}

	reader := newIoprogress(label, res.ContentLength, res.Body)

	_, err = io.Copy(writer, reader)
	if err != nil {
		return "", fmt.Errorf("error copying %s: %v", label, err)
	}

	return res.Header.Get("ETag"), nil
}

func newIoprogress(label string, size int64, rdr io.Reader) io.Reader {
	prefix := "Downloading " + label
	fmtBytesSize := 18
	barSize := int64(80 - len(prefix) - fmtBytesSize)
	bar := ioprogress.DrawTextFormatBarForW(barSize, os.Stderr)
	fmtfunc := func(progress, total int64) string {
		// Content-Length is set to -1 when unknown.
		if total == -1 {
			return fmt.Sprintf(
				"%s: %v of an unknown total size",
				prefix,
				ioprogress.ByteUnitStr(progress),
			)
		}
		return fmt.Sprintf(
			"%s: %s %s",
			prefix,
			bar(progress, total),
			ioprogress.DrawTextFormatBytes(progress, total),
		)
	}

	return &ioprogress.Reader{
		Reader:       rdr,
		Size:         size,
		DrawFunc:     ioprogress.DrawTerminalf(os.Stderr, fmtfunc),
		DrawInterval: time.Second,
	}
}
