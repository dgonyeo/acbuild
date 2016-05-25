// Copyright 2016 The appc Authors
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
	"strings"

	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"

	"github.com/appc/acbuild/registry"
)

// FetchImage will fetch an image represented by imageName into acbuild's
// store.
func (a *ACBuild) FetchImage(imageName string, insecure bool) error {
	reg := registry.Registry{
		DepStoreTarPath:      a.DepStoreTarPath,
		DepStoreExpandedPath: a.DepStoreExpandedPath,
		Insecure:             insecure,
		Debug:                a.Debug,
	}

	name, labels, err := stringToNameAndLabels(imageName)
	if err != nil {
		return err
	}

	return reg.FetchAndRender(*name, labels, 0)
}

// GetImagesInStore will return a list of manifests for each image currently in
// acbuild's store.
func (a *ACBuild) GetImagesInStore() ([]*schema.ImageManifest, error) {
	reg := registry.Registry{
		DepStoreTarPath:      a.DepStoreTarPath,
		DepStoreExpandedPath: a.DepStoreExpandedPath,
	}
	keys, err := reg.GetAllACIs()
	if err != nil {
		return nil, err
	}
	manifests := make([]*schema.ImageManifest, len(keys))
	for i, key := range keys {
		manifest, err := reg.GetImageManifest(key)
		if err != nil {
			return nil, err
		}
		manifests[i] = manifest
	}
	return manifests, nil
}

// ClearStore will empty acbuild's store by deleting it.
func (a *ACBuild) ClearStore() error {
	err := os.RemoveAll(a.DepStoreTarPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	err = os.RemoveAll(a.DepStoreExpandedPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// DeleteImage will delete a single image, represented by imageName, from
// acbuild's store.
func (a *ACBuild) DeleteImage(imageName string) error {
	reg := registry.Registry{
		DepStoreTarPath:      a.DepStoreTarPath,
		DepStoreExpandedPath: a.DepStoreExpandedPath,
	}

	name, labels, err := stringToNameAndLabels(imageName)
	if err != nil {
		return err
	}

	key, err := reg.GetACI(*name, labels)
	if err != nil {
		return err
	}
	return reg.DeleteACI(key)
}

func stringToNameAndLabels(str string) (*types.ACIdentifier, types.Labels, error) {
	tokens := strings.SplitN(str, ":", 2)
	name, err := types.NewACIdentifier(tokens[0])
	if err != nil {
		return nil, nil, err
	}
	var labels types.Labels
	if len(tokens) > 1 {
		labels = append(labels, types.Label{
			Name:  *types.MustACIdentifier("version"),
			Value: tokens[1],
		})
	}
	return name, labels, err
}
