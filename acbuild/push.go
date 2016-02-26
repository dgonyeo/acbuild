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
	"os"
	"path"

	acpush "github.com/appc/acpush/lib"
	"github.com/spf13/cobra"
)

var (
	pushSign      = false
	pushOutput    = ""
	pushOverwrite = false
	pushInsecure  = false
	cmdPush       = &cobra.Command{
		Use:     "push IMAGE_NAME",
		Short:   "Push the image to a registry",
		Long:    "Sets the name of the current ACI, and uploads it to a registry.",
		Example: "acbuild push --sign --output=mynewapp.aci example.com/mynewapp -- --no-default-keyring --keyring ./rkt.gpg",
		Run:     runWrapper(runPush),
	}
)

func init() {
	cmdAcbuild.AddCommand(cmdPush)

	cmdPush.Flags().BoolVar(&pushSign, "sign", false, "sign the resulting ACI")
	cmdPush.Flags().BoolVar(&pushInsecure, "insecure", false, "skip security checks")
	cmdPush.Flags().StringVar(&pushOutput, "output", "", "save the ACI locally at this path")
	cmdPush.Flags().BoolVar(&pushOverwrite, "overwrite", false, "if saving the ACI, whether to overwrite existing files")
}

func runPush(cmd *cobra.Command, args []string) (exit int) {
	if len(args) == 0 {
		cmd.Usage()
		return 1
	}

	if debug && pushOutput != "" {
		stderr("Writing ACI to %s", pushOutput)
	}

	a := newACBuild()

	savingACI := true
	if pushOutput == "" {
		savingACI = false
		pushOutput = path.Join(a.ContextPath, "tmp.aci")
	}

	err := a.Write(pushOutput, overwrite)
	if err != nil {
		stderr("push: %v", err)
		return getErrorCode(err)
	}
	if !savingACI {
		defer os.Remove(pushOutput)
	}

	ascpath := ""
	if pushSign {
		ascpath = pushOutput + ".asc"
		err = a.SignACI(pushOutput, args[1:])
		if err != nil {
			stderr("push: %v", err)
			return getErrorCode(err)
		}
		if !savingACI {
			defer os.Remove(ascpath)
		}
	}

	u := acpush.Uploader{
		Acipath:  pushOutput,
		Ascpath:  ascpath,
		Uri:      args[0],
		Insecure: pushInsecure,
		Debug:    debug,
	}
	err = u.Upload()
	if err != nil {
		stderr("push: %v", err)
		return getErrorCode(err)
	}

	return 0
}
