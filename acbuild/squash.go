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
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/spf13/cobra"
)

var (
	cmdSquash = &cobra.Command{
		Use:     "squash",
		Short:   "Flattens an ACI with its dependencies",
		Example: "acbuild squash",
		Run:     runWrapper(runSquash),
	}
)

func init() {
	cmdAcbuild.AddCommand(cmdSquash)
	cmdSquash.Flags().BoolVar(&insecure, "insecure", false, "Allows fetching dependencies over an unencrypted connection")
}

func runSquash(cmd *cobra.Command, args []string) (exit int) {
	if len(args) != 0 {
		cmd.Usage()
		return 1
	}

	if debug {
		stderr("Squashing ACI")
	}

	err := newACBuild().Squash(insecure)

	if err != nil {
		stderr("squash: %v", err)
		return 1
	}

	return 0
}
