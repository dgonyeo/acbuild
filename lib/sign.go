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
	"os/exec"
)

func (a *ACBuild) SignACI(acipath string, flags []string) error {
	signaturepath := acipath + ".asc"
	if len(flags) == 0 {
		flags = []string{"--armor", "--yes"}
	}
	flags = append(flags, "--output", signaturepath, "--detach-sig", acipath)

	gpgCmd := exec.Command("gpg", flags...)
	gpgCmd.Stdin = os.Stdin
	gpgCmd.Stdout = os.Stdout
	gpgCmd.Stderr = os.Stderr
	return gpgCmd.Run()
}
