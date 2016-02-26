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
	cmdSetName = &cobra.Command{
		Use:     "set-name ACI_NAME",
		Short:   "Set the image name",
		Long:    "Sets the name of the ACI in the manifest",
		Example: "acbuild set-name quay.io/coreos/etcd",
		Run:     runWrapper(runSetName),
	}
)

func init() {
	cmdAcbuild.AddCommand(cmdSetName)
}

func runSetName(cmd *cobra.Command, args []string) (exit int) {
	if len(args) > 1 {
		stderr("set-name: incorrect number of arguments")
		return 1
	}
	if len(args) == 0 {
		cmd.Usage()
		return 1
	}

	if debug {
		stderr("Setting name of ACI to %s", args[0])
	}

	err := newACBuild().SetName(args[0])

	if err != nil {
		stderr("set-name: %v", err)
		return getErrorCode(err)
	}

	return 0
}
