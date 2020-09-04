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
package cmd_scenario

import (
	"fmt"
	"os"

	loader "github.com/geaaru/time-master/pkg/loader"
	scheduler "github.com/geaaru/time-master/pkg/scheduler"
	specs "github.com/geaaru/time-master/pkg/specs"

	tablewriter "github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func NewBuildCommand(config *specs.TimeMasterConfig) *cobra.Command {
	var clients []string
	var activities []string
	var tasksFlags []string

	var cmd = &cobra.Command{
		Use:   "build [scenario]",
		Short: "build of scenarios.",
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("No scenario selected.")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			onlyClosed, _ := cmd.Flags().GetBool("only-closed")
			skipEmptyTasks, _ := cmd.Flags().GetBool("skip-empty-tasks")
			skipPlan, _ := cmd.Flags().GetBool("skip-plan")
			withClientData, _ := cmd.Flags().GetBool("with-client-data")
			preFilter, _ := cmd.Flags().GetBool("pre-filter")
			now, _ := cmd.Flags().GetString("now")
			targetFile, _ := cmd.Flags().GetString("file")

			// Create Instance
			tm := loader.NewTimeMasterInstance(config)

			err := tm.Load()
			if err != nil {
				fmt.Println("Error on load data:" + err.Error() + "\n")
				os.Exit(1)
			}

			sName := args[0]

			scenario, err := tm.GetScenarioByName(sName)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			if now != "" {
				scenario.SetNow(now)
			}

			var sched scheduler.TimeMasterScheduler

			// Create schedule
			switch scenario.Scheduler {

			default:
				sched = scheduler.NewSimpleScheduler(config, scenario)
			}

			tm.InitScheduler(sched)

			opts := scheduler.SchedulerOpts{
				Clients:              clients,
				Activities:           activities,
				FilterPreElaboration: preFilter,
				OnlyClosed:           onlyClosed,
				SkipEmptyTasks:       skipEmptyTasks,
				SkipPlan:             skipPlan,
				ExcludeTaskFlags:     tasksFlags,
			}

			prevision, err := sched.BuildPrevision(opts)
			if err != nil {
				fmt.Println("Error: " + err.Error())
				os.Exit(1)
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetBorders(tablewriter.Border{
				Left:   true,
				Top:    true,
				Right:  true,
				Bottom: true})

			table.SetHeader([]string{
				"Activity",
				"Task",
				"Start Date",
				"End Date",
				"Progress",
			})
			table.SetFooterAlignment(tablewriter.ALIGN_LEFT)
			table.SetColMinWidth(1, 60)
			table.SetColWidth(100)

			for _, t := range prevision.Schedule {

				row := []string{
					t.Activity.Name,
					t.Task.Name,
					t.Period.StartPeriod,
					t.Period.EndPeriod,
					fmt.Sprintf("%02.02f", t.Progress),
				}

				table.Append(row)

			}

			table.Render()

			if !withClientData {

				for idx, _ := range prevision.Schedule {
					prevision.Schedule[idx].Client = nil
				}
			}

			if targetFile != "" {
				err := prevision.Write2File(targetFile)
				if err != nil {
					fmt.Println("Error on write file: " + err.Error())
					os.Exit(1)
				}
			}
		},
	}

	flags := cmd.Flags()
	flags.Bool("only-closed", false, "Show only tasks of closed activities.")
	flags.Bool("skip-empty-tasks", false, "Skip tasks closed without effort.")
	flags.Bool("skip-plan", false, "Avoid simulation and report only available timesheets.")
	flags.Bool("with-client-data", false, "Write also client data on prevision file.")
	flags.Bool("pre-filter", false, "Apply filter before build prevision.")
	flags.StringP("file", "f", "", "Set the file where to write calculate prevision.")
	flags.String("now", "", "Override now value of the scenario in the format YYYY-MM-DD.")

	flags.StringSliceVar(&clients, "client", []string{}, "Filter for client with specified name.")
	flags.StringSliceVar(&tasksFlags, "exclude-flag", []string{}, "Exclude task with specified name.")
	flags.StringSliceVarP(&activities, "activity", "a",
		[]string{}, "Filter for activities with specified name.")

	return cmd
}
