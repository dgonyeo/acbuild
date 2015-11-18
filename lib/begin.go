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
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/appc/acbuild/registry"
	"github.com/appc/acbuild/util"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/aci"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/discovery"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema/types"
)

var (
	placeholdername = "acbuild-unnamed"
)

// Begin will start a new build, storing the untarred ACI the build operates on
// at a.CurrentACIPath. If start is the empty string, the build will begin with
// an empty ACI, otherwise the ACI stored at start will be used at the starting
// point.
func (a *ACBuild) Begin(start string, insecure bool) (err error) {
	_, err = os.Stat(a.ContextPath)
	switch {
	case os.IsNotExist(err):
		break
	case err != nil:
		return err
	default:
		return fmt.Errorf("build already in progress in this working dir")
	}

	err = os.MkdirAll(a.ContextPath, 0755)
	if err != nil {
		return err
	}

	if err = a.lock(); err != nil {
		return err
	}
	defer func() {
		if err1 := a.unlock(); err == nil {
			err = err1
		}
	}()

	defer func() {
		// If there was an error while beginning, we don't want to produce an
		// unexpected build context
		if err != nil {
			os.RemoveAll(a.ContextPath)
		}
	}()

	if start != "" {
		err = os.MkdirAll(a.CurrentACIPath, 0755)
		if err != nil {
			return err
		}
		if start[0] == '.' || start[0] == '/' {
			return a.beginFromLocalImage(start)
		} else {
			return a.beginFromRemoteImage(start, insecure)
		}
	}

	err = os.MkdirAll(path.Join(a.CurrentACIPath, aci.RootfsDir), 0755)
	if err != nil {
		return err
	}

	acid, err := types.NewACIdentifier("acbuild-unnamed")
	if err != nil {
		return err
	}

	archlabel, err := types.NewACIdentifier("arch")
	if err != nil {
		return err
	}

	oslabel, err := types.NewACIdentifier("os")
	if err != nil {
		return err
	}

	manifest := &schema.ImageManifest{
		ACKind:    schema.ImageManifestKind,
		ACVersion: schema.AppContainerVersion,
		Name:      *acid,
		Labels: types.Labels{
			types.Label{
				*archlabel,
				runtime.GOARCH,
			},
			types.Label{
				*oslabel,
				runtime.GOOS,
			},
		},
	}

	manblob, err := manifest.MarshalJSON()
	if err != nil {
		return err
	}

	manfile, err := os.Create(path.Join(a.CurrentACIPath, aci.ManifestFile))
	if err != nil {
		return err
	}

	_, err = manfile.Write(manblob)
	if err != nil {
		return err
	}

	err = manfile.Close()
	if err != nil {
		return err
	}

	return nil
}

func (a *ACBuild) beginFromLocalImage(start string) error {
	finfo, err := os.Stat(start)
	if err == nil {
		if finfo.IsDir() {
			return fmt.Errorf("provided starting ACI is a directory: %s", start)
		}
		return util.ExtractImage(start, a.CurrentACIPath, nil)
	}
	return err
}

func (a *ACBuild) beginFromRemoteImage(start string, insecure bool) error {
	// Check if we're starting with a docker image
	if strings.HasPrefix(start, "docker://") {
		// TODO use docker2aci
		return fmt.Errorf("docker containers are currently unsupported")
	}

	app, err := discovery.NewAppFromString(start)
	if err != nil {
		return err
	}
	labels, err := types.LabelsFromMap(app.Labels)
	if err != nil {
		return err
	}

	err = os.MkdirAll(a.DepStoreTarPath, 0755)
	if err != nil {
		return err
	}

	reg := registry.Registry{
		DepStoreTarPath:      a.DepStoreTarPath,
		DepStoreExpandedPath: a.DepStoreExpandedPath,
		Insecure:             insecure,
		Debug:                a.Debug,
	}

	id, err := reg.Fetch(app.Name, labels, 0, false)
	if err != nil {
		if urlerr, ok := err.(*url.Error); ok {
			if operr, ok := urlerr.Err.(*net.OpError); ok {
				if dnserr, ok := operr.Err.(*net.DNSError); ok {
					if dnserr.Err == "no such host" {
						return fmt.Errorf("unknown host when fetching image, check your connection and local file paths must start with '/' or '.'")
					}
				}
			}
		}
		return err
	}

	return util.ExtractImage(path.Join(a.DepStoreExpandedPath, id), a.CurrentACIPath, nil)
}
