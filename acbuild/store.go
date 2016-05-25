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

package main

import (
	"github.com/spf13/cobra"
)

var (
	flagStoreFetchInsecure = false
	cmdStore               = &cobra.Command{
		Use:   "store [command]",
		Short: "Manage images in acbuild's store",
	}
	cmdStoreFetch = &cobra.Command{
		Use:     "fetch IMAGE",
		Short:   "fetches and renders an image into the store",
		Example: "acbuild store fetch example.com/worker:v1.2.3",
		Run:     runWrapper(runStoreFetch),
	}
	cmdStoreList = &cobra.Command{
		Use:     "list",
		Short:   "List all images in the store",
		Example: "acbuild store list",
		Run:     runWrapper(runStoreList),
	}
	cmdStoreClear = &cobra.Command{
		Use:     "clear",
		Short:   "Deletes all images in the store",
		Example: "acbuild store clear",
		Run:     runWrapper(runStoreClear),
	}
	cmdStoreDelete = &cobra.Command{
		Use:     "rm IMAGE",
		Short:   "Removes an image from the store",
		Example: "acbuild store rm example.com/worker:v1.2.3",
		Run:     runWrapper(runStoreDelete),
	}
)

func init() {
	cmdAcbuild.AddCommand(cmdStore)
	cmdStore.AddCommand(cmdStoreList)
	cmdStore.AddCommand(cmdStoreClear)
	cmdStore.AddCommand(cmdStoreDelete)
	cmdStore.AddCommand(cmdStoreFetch)

	cmdStoreFetch.Flags().BoolVar(&flagStoreFetchInsecure, "insecure", false, "Allows fetching dependencies over http")
}

func runStoreFetch(cmd *cobra.Command, args []string) (exit int) {
	if len(args) != 1 {
		cmd.Usage()
		return 1
	}
	err := newACBuild().FetchImage(args[0], flagStoreFetchInsecure)
	if err != nil {
		stderr("store fetch: %v", err)
		return 1
	}
	return 0
}

func runStoreList(cmd *cobra.Command, args []string) (exit int) {
	if len(args) > 0 {
		cmd.Usage()
		return 1
	}

	images, err := newACBuild().GetImagesInStore()
	if err != nil {
		stderr("store list: %v", err)
		return getErrorCode(err)
	}

	if len(images) == 0 {
		stderr("store list: no images in store")
	}

	for _, image := range images {
		version, ok := image.Labels.Get("version")
		if !ok {
			version = "latest"
		}
		stdout(image.Name.String() + ":" + version)
	}
	return 0
}

func runStoreClear(cmd *cobra.Command, args []string) (exit int) {
	if len(args) > 0 {
		cmd.Usage()
		return 1
	}
	err := newACBuild().ClearStore()
	if err != nil {
		stderr("store clear: %v", err)
		return getErrorCode(err)
	}
	return 0
}

func runStoreDelete(cmd *cobra.Command, args []string) (exit int) {
	if len(args) != 1 {
		cmd.Usage()
		return 1
	}
	err := newACBuild().DeleteImage(args[0])
	if err != nil {
		stderr("store rm: %v", err)
		return 1
	}
	return 0
}
