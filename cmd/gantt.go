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
	"io/ioutil"
	"os"
	"path/filepath"

	gantt "github.com/geaaru/time-master/pkg/gantt"
	specs "github.com/geaaru/time-master/pkg/specs"

	"github.com/spf13/cobra"
)

func loadPrevisionFile(file string) (*specs.ScenarioSchedule, error) {
	fileAbs, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadFile(fileAbs)
	if err != nil {
		return nil, err
	}

	return specs.ScenarioScheduleFromYaml(content, file)
}

func newGanttCommand(config *specs.TimeMasterConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "gantt [OPTIONS]",
		Short: "Produce gantt files.",
		PreRun: func(cmd *cobra.Command, args []string) {

			pFile, _ := cmd.Flags().GetString("prevision")
			if pFile == "" {
				fmt.Println("Mandatory --prevision option missing.")
				os.Exit(1)
			}

		},
		Run: func(cmd *cobra.Command, args []string) {
			pFile, _ := cmd.Flags().GetString("prevision")
			stdOut, _ := cmd.Flags().GetBool("stdout")

			prevision, err := loadPrevisionFile(pFile)
			if err != nil {
				fmt.Println("Error on load prevision file: " + err.Error())
				os.Exit(1)
			}

			producer, err := gantt.NewProducer(config, "frappe")
			if err != nil {
				fmt.Println("Error on create producer: " + err.Error())
				os.Exit(1)
			}

			data, err := producer.Build(prevision)
			if err != nil {
				fmt.Println("Error on produce data: " + err.Error())
				os.Exit(1)
			}

			if stdOut {
				fmt.Println(string(data))
			}

		},
	}

	flags := cmd.Flags()
	flags.String("prevision", "", "Path of the file with the scenario prevision.")
	flags.BoolP("stdout", "o", false, "Write to stdout.")

	return cmd
}
