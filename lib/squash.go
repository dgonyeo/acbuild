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
	"os"
	"path"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema"
	//"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/rkt/pkg/fileutil"
	//"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/rkt/pkg/uid"

	"github.com/appc/acbuild/util"
)

func (a *ACBuild) Squash(insecure bool) (err error) {
	if err = a.lock(); err != nil {
		return err
	}
	defer func() {
		if err1 := a.unlock(); err == nil {
			err = err1
		}
	}()

	err = util.RmAndMkdir(a.OverlayTargetPath)
	if err != nil {
		return err
	}
	defer os.RemoveAll(a.OverlayTargetPath)
	err = os.MkdirAll(a.DepStoreTarPath, 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(a.DepStoreExpandedPath, 0755)
	if err != nil {
		return err
	}

	deps, err := a.renderACI(insecure)
	if err != nil {
		return err
	}

	for _, dep := range deps {
		//err := fileutil.CopyTree(,
		//	a.OverlayTargetPath, uid.NewBlankUidRange())
		err := util.Exec("cp", "-RT", path.Join(a.DepStoreExpandedPath, dep), a.OverlayTargetPath)
		if err != nil {
			return err
		}
	}

	//err = fileutil.CopyTree(a.CurrentACIPath, a.OverlayTargetPath, uid.NewBlankUidRange())
	err = util.Exec("cp", "-RT", path.Join(a.CurrentACIPath), a.OverlayTargetPath)
	if err != nil {
		return err
	}

	err = os.RemoveAll(a.CurrentACIPath)
	if err != nil {
		return err
	}

	err = os.Rename(a.OverlayTargetPath, a.CurrentACIPath)
	if err != nil {
		return err
	}

	fn := func(s *schema.ImageManifest) {
		s.Dependencies = nil
	}
	return util.ModifyManifest(fn, a.CurrentACIPath)

	return nil
}
