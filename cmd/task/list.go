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
package cmd_task

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	loader "github.com/geaaru/time-master/pkg/loader"
	specs "github.com/geaaru/time-master/pkg/specs"
	time "github.com/geaaru/time-master/pkg/time"

	tablewriter "github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func NewListCommand(config *specs.TimeMasterConfig) *cobra.Command {
	var users []string
	var clients []string
	var tasks []string
	var labels []string
	var flagsFilter []string
	var activityLabels []string
	var activityFlags []string

	var cmd = &cobra.Command{
		Use:   "list",
		Short: "list of task.",
		PreRun: func(cmd *cobra.Command, args []string) {
			onlyClosed, _ := cmd.Flags().GetBool("only-closed")
			closed, _ := cmd.Flags().GetBool("closed")
			withMilestone, _ := cmd.Flags().GetBool("with-milestone")
			onlyMilestone, _ := cmd.Flags().GetBool("only-milestone")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			jsonTask, _ := cmd.Flags().GetBool("json-task")
			csvOutput, _ := cmd.Flags().GetBool("csv")

			if onlyClosed && closed {
				fmt.Println("Both option --closed and --only-closed not admitted.")
				os.Exit(1)
			}

			if onlyMilestone && withMilestone {
				fmt.Println("Both option --milestone and --only-milestone not admitted.")
				os.Exit(1)
			}

			if jsonOutput && csvOutput {
				fmt.Println("Both option --csv and --json not admitted.")
				os.Exit(1)
			}

			if jsonTask && !jsonOutput {
				fmt.Println("Use --json-task only with --json")
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

			onlyClosed, _ := cmd.Flags().GetBool("only-closed")
			closed, _ := cmd.Flags().GetBool("closed")
			withEffort, _ := cmd.Flags().GetBool("with-effort")
			withMilestone, _ := cmd.Flags().GetBool("with-milestone")
			onlyMilestone, _ := cmd.Flags().GetBool("only-milestone")
			csvOutput, _ := cmd.Flags().GetBool("csv")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			jsonTask, _ := cmd.Flags().GetBool("json-task")
			minimal, _ := cmd.Flags().GetBool("minimal")
			showWorkHours, _ := cmd.Flags().GetBool("show-work-hours")

			opts := specs.TaskResearch{
				Users:              users,
				Tasks:              tasks,
				Clients:            clients,
				Labels:             labels,
				Flags:              flagsFilter,
				ActivityFlags:      activityFlags,
				ActivityLabels:     activityLabels,
				OnlyClosedActivity: onlyClosed,
				ClosedActivity:     closed,
				OnlyMilestone:      onlyMilestone,
				Milestone:          withMilestone,
				WithEffort:         withEffort,
			}

			res, err := tm.GetTasks(opts)
			if err != nil {
				fmt.Println("Error " + err.Error())
				os.Exit(1)
			}

			researchOpts := specs.TimesheetResearch{
				ByTask:     true,
				IgnoreTime: true,
			}

			rtaMap, err := tm.GetAggregatedTimesheetsMap(researchOpts, "", "", users, tasks)
			if err != nil {
				fmt.Println("Error on elaborate timesheet: " + err.Error())
				os.Exit(1)
			}

			if jsonOutput {

				jsonData := []specs.TaskReport{}

				for _, t := range res {

					tr := specs.TaskReport{
						Name:        t.Name,
						Description: t.Description,
					}

					if jsonTask {
						if minimal {
							tr.Task = t.Clone(true)
						} else {
							tr.Task = t.Clone(false)
						}
					}

					if !minimal {
						durationEffort := ""
						effort := int64(0)
						if t.GetEffort() != "" {
							effort, err = time.ParseDuration(t.GetEffort(), config.GetWork().WorkHours)
							if err != nil {
								fmt.Println(err.Error())
								os.Exit(1)
							}

							durationEffort, err = time.Seconds2Duration(effort)
							if err != nil {
								fmt.Println(err.Error())
								os.Exit(1)
							}
						}

						tr.Effort = durationEffort
						tr.EffortSec = effort
					}

					if showWorkHours {
						rta, _ := rtaMap[t.Name]
						if rta != nil {
							tr.Work = rta.GetDuration()
							tr.WorkSec = rta.GetSeconds()
						}
					}

					jsonData = append(jsonData, tr)
				}

				data, err := json.Marshal(jsonData)
				if err != nil {
					fmt.Println(fmt.Errorf("Error on convert data to json: %s", err.Error()))
					os.Exit(1)
				}
				fmt.Println(string(data))

			} else if csvOutput {

				records := make([][]string, len(res)+1)
				records[0] = []string{"Task", "Description"}
				if showWorkHours {
					records[0] = append(records[0], "Work")
				}

				if !minimal {
					records[0] = append(records[0], "Effort")
				}

				for idx, t := range res {

					records[idx+1] = []string{
						t.Name, t.Description,
					}

					if showWorkHours {
						rta, _ := rtaMap[t.Name]
						work := ""
						if rta != nil {
							work = rta.GetDuration()
						}

						records[idx+1] = append(records[idx+1], work)
					}

					if !minimal {
						if t.GetEffort() != "" {
							effort, err := time.ParseDuration(t.GetEffort(), config.GetWork().WorkHours)
							if err != nil {
								fmt.Println(err.Error())
								os.Exit(1)
							}

							durationEffort, err := time.Seconds2Duration(effort)
							if err != nil {
								fmt.Println(err.Error())
								os.Exit(1)
							}

							records[idx+1] = append(records[idx+1], durationEffort)
						} else {
							records[idx+1] = append(records[idx+1], "")
						}

					}

				}
				w := csv.NewWriter(os.Stdout)
				for _, record := range records {
					if err := w.Write(record); err != nil {
						fmt.Println("error writing record to csv:", err)
						os.Exit(1)
					}
				}

				// Write any buffered data to the underlying writer (standard output).
				w.Flush()

				if err := w.Error(); err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}

			} else {
				table := tablewriter.NewWriter(os.Stdout)
				table.SetBorders(tablewriter.Border{
					Left:   true,
					Top:    true,
					Right:  true,
					Bottom: true})
				headers := []string{"Task", "Description"}

				if showWorkHours {
					headers = append(headers, "Work")
				}
				if !minimal {
					headers = append(headers, "Effort")
				}
				table.SetHeader(headers)
				table.SetFooterAlignment(tablewriter.ALIGN_LEFT)
				table.SetColMinWidth(1, 60)
				table.SetColWidth(100)

				totEffort := int64(0)
				totWork := int64(0)
				for _, t := range res {

					durationEffort := ""
					effort := int64(0)
					if t.GetEffort() != "" {
						effort, err = time.ParseDuration(t.GetEffort(), config.GetWork().WorkHours)
						if err != nil {
							fmt.Println(err.Error())
							os.Exit(1)
						}

						durationEffort, err = time.Seconds2Duration(effort)
						if err != nil {
							fmt.Println(err.Error())
							os.Exit(1)
						}
					}

					row := []string{t.Name, t.Description}
					if showWorkHours {
						rta, _ := rtaMap[t.Name]
						work := ""
						if rta != nil {
							work = rta.GetDuration()
							totWork += rta.GetSeconds()
						}
						row = append(row, work)
					}
					if !minimal {
						row = append(row, durationEffort)
					}

					table.Append(row)
					totEffort += effort
				}

				duration, _ := time.Seconds2Duration(totEffort)
				durationWork, _ := time.Seconds2Duration(totWork)

				footer := []string{fmt.Sprintf("Total (%d)", len(res)), ""}
				if showWorkHours {
					footer = append(footer, durationWork)
				}
				if !minimal {
					footer = append(footer, duration)
				}

				table.SetFooter(footer)
				table.Render()
			}
		},
	}

	flags := cmd.Flags()
	flags.Bool("csv", false, "Print output in CSV format")
	flags.Bool("json", false, "Print output in JSON format")
	flags.Bool("json-task", false, "Add also task object for every entry.")
	flags.Bool("closed", false, "Include tasks of closed activities.")
	flags.Bool("only-closed", false, "Show only tasks of closed activities.")
	flags.Bool("with-effort", false, "Show only tasks with effort")
	flags.Bool("show-work-hours", false, "Show also worked hours")
	flags.Bool("with-milestone", false, "Include tasks of milestone")
	flags.Bool("only-milestone", false, "Show only milestone tasks")
	flags.Bool("minimal", false, "Show only minimal informations.")
	flags.StringSliceVarP(&tasks, "task", "t", []string{},
		"Filter for tasks with regex string.")
	flags.StringSliceVarP(&users, "user", "u", []string{},
		"Filter for tasks allocated for users with regex string.")
	flags.StringSliceVar(&clients, "client", []string{},
		"Filter for clients with regex string.")
	flags.StringSliceVarP(&flagsFilter, "flag", "f", []string{},
		"Filter for tasks that contains flags with regex string.")
	flags.StringSliceVarP(&labels, "label", "l", []string{},
		"Filter for tasks that contains labels with regex string.")
	flags.StringSliceVar(&activityLabels, "activity-label", []string{},
		"Filter for activities that contains labels with regex string.")
	flags.StringSliceVar(&activityFlags, "activity-flag", []string{},
		"Filter for activities that contains flags with regex string.")

	return cmd
}
