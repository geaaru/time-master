/*
Copyright (C) 2020-2021  Daniele Rondina <geaaru@sabayonlinux.org>
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

func retrieveWorkTimeByActivity(tm *loader.TimeMasterInstance, activity,
	scenario, from, to string) (string, int64, float64, float64, error) {

	researchOpts := specs.TimesheetResearch{
		ByActivity: true,
		IgnoreTime: true,
	}

	rtaList, err := tm.GetAggregatedTimesheets(researchOpts, from, to, []string{}, []string{"^" + activity})
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
	var activityNames []string
	var excludeActivityNames []string
	var excludeActivityFlags []string
	var activityLabels []string
	var activityFlags []string
	var clients []string
	var labelsColumn []string

	var cmd = &cobra.Command{
		Use:   "summary [<client-name> [<activity-name>]]",
		Short: "Show a summary of a specific activity.",
		PreRun: func(cmd *cobra.Command, args []string) {
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
			minimal, _ := cmd.Flags().GetBool("minimal")
			labelsInAnd, _ := cmd.Flags().GetBool("labels-in-and")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			csvOutput, _ := cmd.Flags().GetBool("csv")
			scenario, _ := cmd.Flags().GetString("scenario-name")
			scenarioFile, _ := cmd.Flags().GetString("scenario")
			from, _ := cmd.Flags().GetString("from")
			to, _ := cmd.Flags().GetString("to")

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

			if scenario != "" {
				// Assign cost/revenue to ResourceTimesheet
				err := tm.CalculateTimesheetsCostAndRevenue(scenario)
				if err != nil {
					fmt.Println("Error on calculate cost/revenue: " + err.Error())
					os.Exit(1)
				}
			}

			if len(args) > 0 {
				cname := args[0]
				_, err := tm.GetClientByName(cname)
				if err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}

				clients = append(clients, cname)

				if len(args) == 2 {
					aname := args[1]

					_, _, err := tm.GetActivityByName(aname)
					if err != nil {
						fmt.Println(err.Error())
						os.Exit(1)
					}

					activityNames = append(activityNames, aname)
				}
			}

			// Create activity filter
			opts := specs.ActivityResearch{
				ClosedActivity:     closed,
				OnlyClosedActivity: onlyClosed,
				Flags:              activityFlags,
				Labels:             activityLabels,
				Clients:            clients,
				Names:              activityNames,
				ExcludeNames:       excludeActivityNames,
				ExcludeFlags:       excludeActivityFlags,
				LabelsInAnd:        labelsInAnd,
			}

			activities, err := tm.GetActivities(opts)
			if err != nil {
				fmt.Println("Error on retrieve list of the activities: " + err.Error())
				os.Exit(1)
			}

			activitiesReport := []specs.ActivityReport{}

			if len(activities) == 0 {
				if jsonOutput {

					data, err := json.Marshal(activitiesReport)
					if err != nil {
						fmt.Println("Error on convert data to json: " + err.Error())
						os.Exit(1)
					}
					fmt.Println(string(data))

				} else if csvOutput {
					fmt.Println("TODO")
				} else {
					fmt.Println("No activities found")
				}

				os.Exit(0)
			}

			var duration string
			var totEffort, totWork, totOffer int64
			var totCost, totRevenueRate, totProfit float64

			totEffort = int64(0)
			totWork = int64(0)
			totCost = float64(0)
			totRevenueRate = float64(0)
			totProfit = float64(0)

			// Print all activities
			for _, activity := range activities {

				effort, err := activity.GetPlannedEffortTotSecs(config.GetWork().WorkHours)
				if err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}

				// Retrieve work time
				work, workSecs, cost, revenue, err := retrieveWorkTimeByActivity(tm, activity.Name, scenario, from, to)
				if err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}

				profit := float64(activity.Offer) - cost
				if scenario != "" {

					if activity.Offer > 0 && workSecs > 0 {
						totProfit += profit
					} else if activity.IsTimeAndMaterial() {
						profit = revenue - cost
						totProfit += profit
					}
				}

				aReport := specs.NewActivityReport(activity, minimal)

				aReport.SetEffort(effort)
				aReport.SetWorkSecs(workSecs)

				aReport.CalculateWorkPerc()

				if !minimal {
					aReport.SetRevenuePlan(revenue)
					aReport.SetCost(cost)
					aReport.SetProfit(profit)
					aReport.SetWork(work)
					aReport.CalculateProfitPerc()
					aReport.CalculateDuration()

					if scenario != "" {
						// Calculate business progress
						bProgress, err := tm.CalculateActivityBusinessProgress(activity.Name)
						if err != nil {
							fmt.Println(err.Error())
							os.Exit(1)
						}
						bprogressPerc := fmt.Sprintf("%02.02f", bProgress)

						if activity.Closed {
							// I consider closed the job
							aReport.SetBusinessProgressPerc("100.0")
						} else {
							aReport.SetBusinessProgressPerc(bprogressPerc)
						}
					}
				} else {
					// Reset effort/secs
					aReport.SetEffort(0)
					aReport.SetWorkSecs(0)
					aReport.SetWorkPerc("")

					if scenario != "" {
						// Calculate business progress
						bProgress, err := tm.CalculateActivityBusinessProgress(activity.Name)
						if err != nil {
							fmt.Println(err.Error())
							os.Exit(1)
						}
						bprogressPerc := fmt.Sprintf("%02.02f", bProgress)

						if activity.Closed {
							// I consider closed the job
							aReport.SetBusinessProgressPerc("100.0")
						} else {
							aReport.SetBusinessProgressPerc(bprogressPerc)
						}
					}
				}

				activitiesReport = append(activitiesReport, *aReport)

				totEffort += effort
				totOffer += activity.Offer
				totWork += workSecs
				totCost += cost
				totRevenueRate += revenue
			}

			if jsonOutput {
				data, err := json.Marshal(activitiesReport)
				if err != nil {
					fmt.Println("Error on convert activities to json: " + err.Error())
					os.Exit(1)
				}
				fmt.Println(string(data))
			} else {
				var table *tablewriter.Table
				records := make([][]string, len(activitiesReport)+1)

				headers := []string{
					"Name", "Description",
				}
				if !minimal {
					headers = append(headers, []string{"% (of Plan)", "# Tasks", "Work", "Effort"}...)
					if scenario != "" {
						headers = append(headers, []string{
							"Business Progress",
							"Cost", "Offer", "Revenue Plan", "Profit", "% Profit",
						}...)
					}
				}

				if len(labelsColumn) > 0 {
					for _, l := range labelsColumn {
						headers = append(headers, l)
					}
				}

				if !csvOutput {

					table = tablewriter.NewWriter(os.Stdout)
					table.SetBorders(tablewriter.Border{
						Left:   true,
						Top:    true,
						Right:  true,
						Bottom: true})

					table.SetHeader(headers)
					table.SetFooterAlignment(tablewriter.ALIGN_LEFT)
					table.SetColMinWidth(1, 60)
					table.SetColWidth(100)

					duration, _ = time.Seconds2Duration(totEffort)
					durationWork, _ := time.Seconds2Duration(totWork)
					footers := []string{
						fmt.Sprintf("Total (%d)", len(activities)),
						"",
						"",
					}

					if !minimal {
						footers = append(footers, []string{"", durationWork, duration}...)
						if scenario != "" {
							profit_perc := fmt.Sprintf("%02.02f", ((float64(totProfit) * 100) / float64(totCost+totProfit)))
							footers = append(footers, []string{
								durationWork,
								fmt.Sprintf("%02.02f", totCost),
								fmt.Sprintf("%d", totOffer),
								fmt.Sprintf("%02.02f", totRevenueRate),
								fmt.Sprintf("%02.02f", totProfit),
								fmt.Sprintf("%s", profit_perc),
							}...)
						}
					}

					if len(labelsColumn) > 0 {
						for range labelsColumn {
							footers = append(footers, "")
						}
					}

					table.SetFooter(footers)

				} else {
					records[0] = headers
				}

				for idx, activity := range activitiesReport {

					row := []string{
						activity.Name,
						activity.Description,
						activity.WorkPerc,
					}

					if !minimal {
						row = append(row, []string{
							fmt.Sprintf("%d", len(activity.Tasks)),
							activity.Work,
							activity.GetDuration(),
						}...)

						if scenario != "" {

							row = append(row, activity.BusinessProgressPerc)
							row = append(row, fmt.Sprintf("%02.02f", activity.Cost))
							row = append(row, fmt.Sprintf("%d", activity.Offer))
							row = append(row, fmt.Sprintf("%02.02f", activity.RevenuePlan))
							if activity.Offer > 0 && activity.WorkSecs > 0 {
								row = append(row, fmt.Sprintf("%02.02f", activity.Profit))
							} else if activity.IsTimeAndMaterial() {
								row = append(row, fmt.Sprintf("%02.02f", activity.Profit))
							} else {
								row = append(row, "0")
							}
							row = append(row, fmt.Sprintf("%s", activity.ProfitPerc))
						}
					}

					if len(labelsColumn) > 0 {
						for _, l := range labelsColumn {
							row = append(row, activity.GetLabelValue(l, ""))
						}
					}

					if csvOutput {
						records[idx+1] = row
					} else {
						table.Append(row)
					}

				}

				if csvOutput {
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
					table.Render()
				}

			}

		},
	}

	flags := cmd.Flags()
	flags.Bool("minimal", false, "Show only minimal data on report.")
	flags.Bool("closed", false, "Include closed activities.")
	flags.Bool("json", false, "Print output in JSON format.")
	flags.Bool("csv", false, "Print output in CSV format.")
	flags.Bool("only-closed", false, "Show only closed activities.")
	flags.Bool("labels-in-and", false, "Filter labels in AND. Default match is in OR.")
	flags.String("scenario-name", "", "Specify scenario name for cost/revenue.")
	flags.String("scenario", "", "Specify path of the scenario prevision to load.")

	flags.StringSliceVar(&clients, "client", []string{}, "Filter for client with specified name.")
	flags.StringSliceVar(&labelsColumn, "label-column", []string{}, "Add label value to output table/CSV")
	flags.StringSliceVarP(&activityNames, "activity", "a",
		[]string{}, "Filter for activities with specified name.")

	flags.StringSliceVar(&activityFlags, "flag", []string{},
		"Filter for activities with specificied flag.")
	flags.StringSliceVar(&activityLabels, "label", []string{},
		"Filter for activities with specificied label.")
	flags.StringSliceVar(&excludeActivityNames, "exclude-activity", []string{},
		"Exclude activities from report.")
	flags.StringSliceVar(&excludeActivityFlags, "exclude-aflag", []string{},
		"Exclude activities with matched flag from report.")

	flags.String("from", "", "Specify from date in format YYYY-MM-DD.")
	flags.String("to", "", "Specify to date in format YYYY-MM-DD.")

	return cmd
}
