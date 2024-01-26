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

	gantt "github.com/geaaru/time-master/pkg/gantt"
	specs "github.com/geaaru/time-master/pkg/specs"

	"github.com/spf13/cobra"
)

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
			toFile, _ := cmd.Flags().GetString("to")
			stdOut, _ := cmd.Flags().GetBool("stdout")
			byEndTime, _ := cmd.Flags().GetBool("by-endtime")
			showActivity, _ := cmd.Flags().GetBool("show-activity")

			prevision, err := specs.ScenarioScheduleFromFile(pFile)
			if err != nil {
				fmt.Println("Error on load prevision file: " + err.Error())
				os.Exit(1)
			}

			producer, err := gantt.NewProducer(config, "frappe")
			if err != nil {
				fmt.Println("Error on create producer: " + err.Error())
				os.Exit(1)
			}

			opts := gantt.ProducerOpts{
				ShowActivityOnTasks: showActivity,
				OrderByEndTime:      byEndTime,
			}

			data, err := producer.Build(prevision, opts)
			if err != nil {
				fmt.Println("Error on produce data: " + err.Error())
				os.Exit(1)
			}

			if toFile != "" {
				err := ioutil.WriteFile(toFile, data, 0644)
				if err != nil {
					fmt.Println("Error on write data on file: " + err.Error())
					os.Exit(1)
				}
			}

			if stdOut {
				fmt.Println(string(data))
			}

		},
	}

	flags := cmd.Flags()
	flags.String("prevision", "", "Path of the file with the scenario prevision.")
	flags.String("to", "", "Path of the file where write gantt data.")
	flags.BoolP("stdout", "o", false, "Write to stdout.")
	flags.Bool("show-activity", false,
		"Add activity name as prefix of task description")
	flags.Bool("by-endtime", false, "Order tasks by end time instead of start time.")

	return cmd
}
