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
package cmd_resource

import (
	"fmt"
	"os"
	"sort"

	loader "github.com/geaaru/time-master/pkg/loader"
	specs "github.com/geaaru/time-master/pkg/specs"
	time "github.com/geaaru/time-master/pkg/time"

	tablewriter "github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func NewTimesheetCommand(config *specs.TimeMasterConfig) *cobra.Command {
	var tasks []string

	var cmd = &cobra.Command{
		Use:   "timesheet [user]",
		Short: "Show resources timesheets.",
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("Missing user")
				os.Exit(1)
			}

			byTask, _ := cmd.Flags().GetBool("by-tasks")
			byActivity, _ := cmd.Flags().GetBool("by-activities")
			ignoreTime, _ := cmd.Flags().GetBool("ignore-time")
			if ignoreTime && !byTask {
				fmt.Println("With ignore-time it's needed by-tasks")
				os.Exit(1)
			}

			if byTask && byActivity {
				fmt.Println("Both options --by-tasks and --by-activities not admitted.")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			monthly, _ := cmd.Flags().GetBool("monthly")
			byTask, _ := cmd.Flags().GetBool("by-tasks")
			byActivity, _ := cmd.Flags().GetBool("by-activities")
			ignoreTime, _ := cmd.Flags().GetBool("ignore-time")
			from, _ := cmd.Flags().GetString("from")
			to, _ := cmd.Flags().GetString("to")

			// Create Instance
			tm := loader.NewTimeMasterInstance(config)

			err := tm.Load()
			if err != nil {
				fmt.Println("Error on load data:" + err.Error() + "\n")
				os.Exit(1)
			}

			researchOpts := specs.TimesheetResearch{
				ByUser:     true,
				ByTask:     byTask,
				ByActivity: byActivity,
				Monthly:    monthly,
				IgnoreTime: ignoreTime,
			}

			user := args[0]
			rtaList, err := tm.GetAggregatedTimesheets(researchOpts, from, to, []string{user}, tasks)
			if err != nil {
				fmt.Println("Error: " + err.Error())
				os.Exit(1)
			}
			if len(*rtaList) == 0 {
				os.Exit(0)
			}

			dates := []string{}
			rtaMap := make(map[string]specs.ResourceTsAggregated, 0)
			for _, rta := range *rtaList {
				var key string
				if !ignoreTime {
					key = rta.Period.StartPeriod
				}
				if byTask || byActivity {
					key += " - " + rta.Task
				}
				dates = append(dates, key)
				rtaMap[key] = rta
			}

			sort.Strings(dates)

			var totEffort int64 = 0
			var dateStr string

			table := tablewriter.NewWriter(os.Stdout)
			table.SetBorders(tablewriter.Border{
				Left:   true,
				Top:    true,
				Right:  true,
				Bottom: true})

			header := []string{}
			if !ignoreTime {
				if monthly {
					dateStr = "Month"
				} else {
					dateStr = "Date"
				}
				header = append(header, dateStr)
			}

			if byTask {
				header = append(header, "Task")
			} else if byActivity {
				header = append(header, "Activity")
			}

			header = append(header, "Effort")
			table.SetHeader(header)
			table.SetFooterAlignment(tablewriter.ALIGN_LEFT)

			for _, d := range dates {
				rta := rtaMap[d]
				totEffort += rta.GetSeconds()
				row := []string{}
				if !ignoreTime {
					row = append(row, rta.Period.StartPeriod)
				}
				if byTask {
					row = append(row, rta.Task)
				}
				row = append(row, rta.GetDuration())

				table.Append(row)
			}

			duration, err := time.Seconds2Duration(totEffort)
			if (byTask || byActivity) && !ignoreTime {
				table.SetFooter([]string{
					"Total",
					"",
					duration,
				})
			} else {
				table.SetFooter([]string{
					"Total",
					duration,
				})
			}

			table.Render()
		},
	}

	flags := cmd.Flags()
	flags.BoolP("monthly", "m", false, "Timesheets aggregated for month instead of day.")
	flags.Bool("by-tasks", false, "Timesheets aggregated for task.")
	flags.Bool("by-activities", false, "Timesheets aggregated for activities.")
	flags.Bool("ignore-time", false, "Timesheets aggregated without monthly/daily aggregation.")
	flags.String("from", "", "Specify from date in format YYYY-MM-DD.")
	flags.String("to", "", "Specify to date in format YYYY-MM-DD.")
	flags.StringSliceVarP(&tasks, "tasks", "t", []string{},
		"Filter for tasks with regex string.")

	return cmd
}
