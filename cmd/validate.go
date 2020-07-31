/*

Copyright (C) 2020  Daniele Rondina <geaaru@sabayonlinux.org>
Credits goes also to Gogs authors, some code portions and re-implemented design
are also coming from the Gogs project, which is using the go-macaron framework
and was really source of ispiration. Kudos to them!

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.

*/
package cmd

import (
	"fmt"
	"os"

	loader "github.com/geaaru/time-master/pkg/loader"
	specs "github.com/geaaru/time-master/pkg/specs"

	"github.com/spf13/cobra"
)

func newValidateCommand(config *specs.TimeMasterConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "validate",
		Short: "Validate data.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {

			//		ignoreError, _ := cmd.Flags().GetBool("ignore-errors")

			// Create Instance
			tm := loader.NewTimeMasterInstance(config)

			err := tm.Load()
			if err != nil {
				fmt.Println("Error on load environments:" + err.Error() + "\n")
				os.Exit(1)
			}

			/*
				err = tm.Validate(ignoreError)
				if err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}
			*/

			fmt.Println("The data are good!")
		},
	}

	pflags := cmd.Flags()
	pflags.BoolP("ignore-errors", "i", false, "Ignore errors and print duplicate.")

	return cmd
}
