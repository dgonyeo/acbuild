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
	"github.com/appc/acbuild/util"

	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
)

func removeMount(name types.ACName) func(*schema.ImageManifest) error {
	return func(s *schema.ImageManifest) error {
		if s.App == nil {
			return ErrNotFound
		}
		foundOne := false
		for i := len(s.App.MountPoints) - 1; i >= 0; i-- {
			if s.App.MountPoints[i].Name == name {
				foundOne = true
				s.App.MountPoints = append(
					s.App.MountPoints[:i],
					s.App.MountPoints[i+1:]...)
			}
		}
		if !foundOne {
			return ErrNotFound
		}
		return nil
	}
}

// AddMount will add a mount point with the given name and path to the untarred
// ACI stored at a.CurrentACIPath. If the mount point already exists its value
// will be updated to the new value. readOnly signifies whether or not the
// mount point should be read only.
func (a *ACBuild) AddMount(name, path string, readOnly bool) (err error) {
	if err = a.lock(); err != nil {
		return err
	}
	defer func() {
		if err1 := a.unlock(); err == nil {
			err = err1
		}
	}()

	acn, err := types.NewACName(name)
	if err != nil {
		return err
	}

	fn := func(s *schema.ImageManifest) error {
		removeMount(*acn)(s)
		if s.App == nil {
			s.App = newManifestApp()
		}
		s.App.MountPoints = append(s.App.MountPoints,
			types.MountPoint{
				Name:     *acn,
				Path:     path,
				ReadOnly: readOnly,
			})
		return nil
	}
	return util.ModifyManifest(fn, a.CurrentACIPath)
}

// RemoveMount will remove the mount point with the given name from the
// untarred ACI stored at a.CurrentACIPath
func (a *ACBuild) RemoveMount(name string) (err error) {
	if err = a.lock(); err != nil {
		return err
	}
	defer func() {
		if err1 := a.unlock(); err == nil {
			err = err1
		}
	}()

	acn, err := types.NewACName(name)
	if err != nil {
		return err
	}

	return util.ModifyManifest(removeMount(*acn), a.CurrentACIPath)
}
