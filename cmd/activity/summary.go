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
package cmd_activity

import (
	"fmt"
	"os"

	loader "github.com/geaaru/time-master/pkg/loader"
	specs "github.com/geaaru/time-master/pkg/specs"
	time "github.com/geaaru/time-master/pkg/time"

	tablewriter "github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func NewSummaryCommand(config *specs.TimeMasterConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "summary <client-name> <activity-name>",
		Short: "Show a summary of a specific activity.",
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("Missing client name")
				os.Exit(1)
			}
			if len(args) > 2 {
				fmt.Println("Too many arguments")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			// Create Instance
			tm := loader.NewTimeMasterInstance(config)

			err := tm.Load()
			if err != nil {
				fmt.Println("Error on load data:" + err.Error() + "\n")
				os.Exit(1)
			}

			cname := args[0]
			client, err := tm.GetClientByName(cname)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			var duration string
			var totEffort int64

			totEffort = 0
			table := tablewriter.NewWriter(os.Stdout)
			table.SetBorders(tablewriter.Border{
				Left:   true,
				Top:    true,
				Right:  true,
				Bottom: true})
			table.SetHeader([]string{"Name", "Description", "# Tasks", "Effort"})
			table.SetFooterAlignment(tablewriter.ALIGN_LEFT)
			table.SetColMinWidth(1, 60)
			table.SetColWidth(100)

			if len(args) == 2 {
				aname := args[1]

				activity, err := client.GetActivityByName(aname)
				if err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}

				effort, err := activity.GetPlannedEffortTotSecs(config.GetWork().WorkHours)
				if err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}
				duration, err = time.Seconds2Duration(effort)

				table.Append([]string{
					activity.Name,
					activity.Description,
					fmt.Sprintf("%d", len(activity.Tasks)),
					duration,
				})

				totEffort += effort
			} else {

				// Print all activities
				for _, activity := range *client.GetActivities() {

					effort, err := activity.GetPlannedEffortTotSecs(config.GetWork().WorkHours)
					if err != nil {
						fmt.Println(err.Error())
						os.Exit(1)
					}
					duration, err = time.Seconds2Duration(effort)

					table.Append([]string{
						activity.Name,
						activity.Description,
						fmt.Sprintf("%d", len(activity.Tasks)),
						duration,
					})

					totEffort += effort
				}

			}

			duration, err = time.Seconds2Duration(totEffort)
			table.SetFooter([]string{
				"Total",
				"",
				"",
				duration,
			})

			table.Render()
		},
	}

	return cmd
}
