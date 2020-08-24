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
package cmd_timesheet

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

func NewShowCommand(config *specs.TimeMasterConfig) *cobra.Command {
	var tasks []string
	var users []string

	var cmd = &cobra.Command{
		Use:   "show",
		Short: "Show timesheets summary.",
		PreRun: func(cmd *cobra.Command, args []string) {
			byTask, _ := cmd.Flags().GetBool("by-tasks")
			byUser, _ := cmd.Flags().GetBool("by-users")
			byActivity, _ := cmd.Flags().GetBool("by-activities")
			ignoreTime, _ := cmd.Flags().GetBool("ignore-time")
			if ignoreTime && !byTask && !byUser && !byActivity {
				fmt.Println(
					"With ignore-time it's needed by-tasks or by-users or by-activities")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			monthly, _ := cmd.Flags().GetBool("monthly")
			byTask, _ := cmd.Flags().GetBool("by-tasks")
			byUser, _ := cmd.Flags().GetBool("by-users")
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
				ByUser:     byUser,
				ByTask:     byTask,
				ByActivity: byActivity,
				Monthly:    monthly,
				IgnoreTime: ignoreTime,
			}

			rtaList, err := tm.GetAggregatedTimesheets(researchOpts, from, to, users, tasks)
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
				if byUser {
					key += " - " + rta.User
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

			headers := []string{}
			footer := []string{"Total"}

			if !ignoreTime {
				if monthly {
					dateStr = "Month"
				} else {
					dateStr = "Date"
				}
				headers = append(headers, dateStr)
			}
			if byUser {
				headers = append(headers, "User")
				if !ignoreTime {
					footer = append(footer, "")
				}
			}

			if byTask {
				headers = append(headers, "Task")
				if (ignoreTime && byUser) || !ignoreTime {
					footer = append(footer, "")
				}
			} else if byActivity {
				headers = append(headers, "Activity")
				if (ignoreTime && byUser) || !ignoreTime {
					footer = append(footer, "")
				}
			}

			headers = append(headers, "Effort")
			table.SetHeader(headers)
			table.SetFooterAlignment(tablewriter.ALIGN_LEFT)

			for _, d := range dates {
				rta := rtaMap[d]
				totEffort += rta.GetSeconds()

				row := []string{}

				if !ignoreTime {
					row = append(row, rta.Period.StartPeriod)
				}
				if byUser {
					row = append(row, rta.User)
				}
				if byTask || byActivity {
					row = append(row, rta.Task)
				}
				row = append(row, rta.GetDuration())
				table.Append(row)
			}

			duration, err := time.Seconds2Duration(totEffort)
			footer = append(footer, duration)
			table.SetFooter(footer)
			table.Render()
		},
	}

	flags := cmd.Flags()
	flags.BoolP("monthly", "m", false, "Timesheets aggregated for month instead of day.")
	flags.Bool("by-tasks", false, "Timesheets aggregated for tasks.")
	flags.Bool("by-users", false, "Timesheets aggregated for users.")
	flags.Bool("by-activities", false, "Timesheets aggregated for activities.")
	flags.Bool("ignore-time", false,
		"Timesheets aggregated without monthly/daily aggregation.")
	flags.String("from", "", "Specify from date in format YYYY-MM-DD.")
	flags.String("to", "", "Specify to date in format YYYY-MM-DD.")
	flags.StringSliceVarP(&tasks, "tasks", "t", []string{}, "Filter for tasks with regex string.")
	flags.StringSliceVarP(&users, "users", "u", []string{}, "Filter for users.")

	return cmd
}
