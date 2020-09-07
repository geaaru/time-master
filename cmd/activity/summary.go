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

func retrieveWorkTimeByActivity(tm *loader.TimeMasterInstance, activity, scenario string) (string, int64, float64, float64, error) {

	researchOpts := specs.TimesheetResearch{
		ByActivity: true,
		IgnoreTime: true,
	}

	if scenario != "" {
		// Assign cost/revenue to ResourceTimesheet
		err := tm.CalculateTimesheetsCostAndRevenue(scenario)
		if err != nil {
			return "", 0, 0, 0, err
		}
	}

	rtaList, err := tm.GetAggregatedTimesheets(researchOpts, "", "", []string{}, []string{activity})
	if err != nil {
		return "", 0, 0, 0, err
	}

	if len(*rtaList) > 0 {
		return (*rtaList)[0].GetDuration(), (*rtaList)[0].GetSeconds(),
			(*rtaList)[0].GetCost(), (*rtaList)[0].GetRevenue(), nil
	}
	return "0", 0, 0, 0, nil
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
			scenario, _ := cmd.Flags().GetString("scenario-name")
			scenarioFile, _ := cmd.Flags().GetString("scenario")

			// Create Instance
			tm := loader.NewTimeMasterInstance(config)

			err := tm.Load()
			if err != nil {
				fmt.Println("Error on load data:" + err.Error() + "\n")
				os.Exit(1)
			}

			if scenarioFile != "" {
				prevision, err := specs.ScenarioScheduleFromFile(scenarioFile)
				if err != nil {
					fmt.Println("Error on load scenario file: " + err.Error())
					os.Exit(1)
				}

				tm.SetAgendaTimesheets([]specs.AgendaTimesheets{
					*prevision.GetAllResourceTimesheets(),
				})
			}

			cname := args[0]
			client, err := tm.GetClientByName(cname)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			var duration string
			var totEffort, totWork, totOffer int64
			var totCost, totRevenueRate, totProfit float64

			totEffort = int64(0)
			totWork = int64(0)
			totCost = float64(0)
			totRevenueRate = float64(0)
			totProfit = float64(0)

			table := tablewriter.NewWriter(os.Stdout)
			table.SetBorders(tablewriter.Border{
				Left:   true,
				Top:    true,
				Right:  true,
				Bottom: true})
			headers := []string{"Name", "Description", "# Tasks", "% (of Plan)", "Work", "Effort"}
			if scenario != "" {
				headers = append(headers, []string{"Cost", "Offer", "Revenue on Rate", "Profit"}...)
			}

			table.SetHeader(headers)
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

				nActivity = 1

				effort, err := activity.GetPlannedEffortTotSecs(config.GetWork().WorkHours)
				if err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}
				duration, err = time.Seconds2Duration(effort)

				// Retrieve work time
				work, workSecs, cost, revenue, err := retrieveWorkTimeByActivity(tm, activity.Name, scenario)
				if err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}

				perc := ""
				if workSecs > 0 && effort > 0 {
					perc = fmt.Sprintf("%02.02f", (float64(workSecs)/float64(effort))*100)
				}

				row := []string{
					activity.Name,
					activity.Description,
					fmt.Sprintf("%d", len(activity.Tasks)),
					perc,
					work,
					duration,
				}

				profit := float64(activity.Offer) - cost
				if scenario != "" {
					row = append(row, fmt.Sprintf("%02.02f", cost))
					row = append(row, fmt.Sprintf("%d", activity.Offer))
					row = append(row, fmt.Sprintf("%02.02f", revenue))
					if activity.Offer > 0 && workSecs > 0 {
						profit_perc := fmt.Sprintf("%02.02f", ((float64(profit) * 100) / float64(activity.Offer)))
						row = append(row, fmt.Sprintf("%02.02f (%s)", profit, profit_perc))
						totProfit += profit
					} else if activity.IsTimeAndMaterial() {
						profit := revenue - cost
						profit_perc := fmt.Sprintf("%02.02f", ((float64(profit) * 100) / float64(revenue)))
						row = append(row, fmt.Sprintf("%02.02f (%s)", profit, profit_perc))
						totProfit += profit
					} else {
						row = append(row, "0")
					}
				}

				table.Append(row)

				totEffort += effort
				totWork += workSecs
				totOffer += activity.Offer
				totCost += cost
				totRevenueRate += revenue
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
					work, workSecs, cost, revenue, err := retrieveWorkTimeByActivity(tm, activity.Name, scenario)
					if err != nil {
						fmt.Println(err.Error())
						os.Exit(1)
					}

					perc := ""
					if workSecs > 0 && effort > 0 {
						perc = fmt.Sprintf("%02.02f", (float64(workSecs)/float64(effort))*100)
					}

					row := []string{
						activity.Name,
						activity.Description,
						fmt.Sprintf("%d", len(activity.Tasks)),
						perc,
						work,
						duration,
					}

					profit := float64(activity.Offer) - cost
					if scenario != "" {

						row = append(row, fmt.Sprintf("%02.02f", cost))
						row = append(row, fmt.Sprintf("%d", activity.Offer))
						row = append(row, fmt.Sprintf("%02.02f", revenue))
						if activity.Offer > 0 && workSecs > 0 {

							profit_perc := fmt.Sprintf("%02.02f", ((float64(profit) * 100) / float64(activity.Offer)))
							row = append(row, fmt.Sprintf("%02.02f (%s)", profit, profit_perc))
							totProfit += profit
						} else if activity.IsTimeAndMaterial() {
							profit := revenue - cost
							profit_perc := fmt.Sprintf("%02.02f", ((float64(profit) * 100) / float64(revenue)))
							row = append(row, fmt.Sprintf("%02.02f (%s)", profit, profit_perc))
							totProfit += profit

						} else {
							row = append(row, "0")
						}
					}

					table.Append(row)

					nActivity += 1
					totEffort += effort
					totOffer += activity.Offer
					totWork += workSecs
					totCost += cost
					totRevenueRate += revenue
				}

			}

			duration, err = time.Seconds2Duration(totEffort)
			durationWork, err := time.Seconds2Duration(totWork)

			if nActivity == 0 {
				fmt.Println("No activities found")
			} else {

				footers := []string{
					fmt.Sprintf("Total (%d)", nActivity),
					"",
					"",
					"",
					durationWork,
					duration,
				}

				if scenario != "" {
					profit_perc := fmt.Sprintf("%02.02f", ((float64(totProfit) * 100) / float64(totCost+totProfit)))
					footers = append(footers, []string{
						fmt.Sprintf("%02.02f", totCost),
						fmt.Sprintf("%d", totOffer),
						fmt.Sprintf("%02.02f", totRevenueRate),
						fmt.Sprintf("%02.02f (%s)", totProfit, profit_perc),
					}...)
				}

				table.SetFooter(footers)

				table.Render()
			}
		},
	}

	flags := cmd.Flags()
	flags.Bool("closed", false, "Include closed activities.")
	flags.Bool("only-closed", false, "Show only closed activities.")
	flags.String("scenario-name", "", "Specify scenario name for cost/revenue.")
	flags.String("scenario", "", "Specify path of the scenario prevision to load.")

	return cmd
}
