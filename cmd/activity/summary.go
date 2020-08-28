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

func retrieveWorkTimeByActivity(tm *loader.TimeMasterInstance, activity string) (string, int64, error) {

	researchOpts := specs.TimesheetResearch{
		ByActivity: true,
		IgnoreTime: true,
	}

	rtaList, err := tm.GetAggregatedTimesheets(researchOpts, "", "", []string{}, []string{activity})
	if err != nil {
		return "", 0, err
	}

	if len(*rtaList) > 0 {
		return (*rtaList)[0].GetDuration(), (*rtaList)[0].GetSeconds(), nil
	}
	return "0", 0, nil
}

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

			onlyClosed, _ := cmd.Flags().GetBool("only-closed")
			closed, _ := cmd.Flags().GetBool("closed")
			if onlyClosed && closed {
				fmt.Println("Both option --closed and --only-closed not admitted.")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			onlyClosed, _ := cmd.Flags().GetBool("only-closed")
			closed, _ := cmd.Flags().GetBool("closed")

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
			var totWork int64

			totEffort = 0
			totWork = 0
			table := tablewriter.NewWriter(os.Stdout)
			table.SetBorders(tablewriter.Border{
				Left:   true,
				Top:    true,
				Right:  true,
				Bottom: true})
			table.SetHeader([]string{"Name", "Description", "# Tasks", "% (of Plan)", "Work", "Effort"})
			table.SetFooterAlignment(tablewriter.ALIGN_LEFT)
			table.SetColMinWidth(1, 60)
			table.SetColWidth(100)
			nActivity := 0

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

				// Retrieve work time
				work, workSecs, err := retrieveWorkTimeByActivity(tm, activity.Name)
				if err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}

				perc := ""
				if workSecs > 0 {
					perc = fmt.Sprintf("%02.02f", (float64(workSecs)/float64(effort))*100)
				}

				table.Append([]string{
					activity.Name,
					activity.Description,
					fmt.Sprintf("%d", len(activity.Tasks)),
					perc,
					work,
					duration,
				})

				totEffort += effort
				totWork += workSecs
			} else {

				// Print all activities
				for _, activity := range *client.GetActivities() {

					if onlyClosed {
						if !activity.IsClosed() {
							continue
						}
					} else if !closed && activity.IsClosed() {
						continue
					}

					effort, err := activity.GetPlannedEffortTotSecs(config.GetWork().WorkHours)
					if err != nil {
						fmt.Println(err.Error())
						os.Exit(1)
					}

					if effort > 0 {
						duration, err = time.Seconds2Duration(effort)
					} else {
						duration = ""
					}

					// Retrieve work time
					work, workSecs, err := retrieveWorkTimeByActivity(tm, activity.Name)
					if err != nil {
						fmt.Println(err.Error())
						os.Exit(1)
					}

					perc := ""
					if workSecs > 0 {
						perc = fmt.Sprintf("%02.02f", (float64(workSecs)/float64(effort))*100)
					} else {
						work = ""
					}

					table.Append([]string{
						activity.Name,
						activity.Description,
						fmt.Sprintf("%d", len(activity.Tasks)),
						perc,
						work,
						duration,
					})

					nActivity += 1
					totEffort += effort
					totWork += workSecs
				}

			}

			duration, err = time.Seconds2Duration(totEffort)
			durationWork, err := time.Seconds2Duration(totWork)

			if nActivity == 0 {
				fmt.Sprintf("No activities found")
			} else {
				table.SetFooter([]string{
					fmt.Sprintf("Total (%d)", nActivity),
					"",
					"",
					"",
					durationWork,
					duration,
				})

				table.Render()
			}
		},
	}

	flags := cmd.Flags()
	flags.Bool("closed", false, "Include closed activities.")
	flags.Bool("only-closed", false, "Show only closed activities.")

	return cmd
}
