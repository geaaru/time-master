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
package cmd_change_request

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	loader "github.com/geaaru/time-master/pkg/loader"
	specs "github.com/geaaru/time-master/pkg/specs"

	tablewriter "github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func NewListCommand(config *specs.TimeMasterConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "list of Change Requests.",
		Run: func(cmd *cobra.Command, args []string) {

			// Create Instance
			tm := loader.NewTimeMasterInstance(config)

			err := tm.Load()
			if err != nil {
				fmt.Println("Error on load data:" + err.Error() + "\n")
				os.Exit(1)
			}

			csvOutput, _ := cmd.Flags().GetBool("csv")
			jsonOutput, _ := cmd.Flags().GetBool("json")

			res := []specs.ChangeRequestReport{}

			for _, c := range *tm.GetClients() {
				for _, a := range *c.GetActivities() {
					crs := a.GetChangeRequests()
					if crs != nil && len(*crs) > 0 {
						for _, cr := range *crs {
							res = append(res, *specs.NewChangeRequestReport(&cr, a.Name))
						}
					}
				}
			}

			if csvOutput {

				w := csv.NewWriter(os.Stdout)
				records := make([][]string, len(res)+1)
				records[0] = []string{
					"Activity", "CR", "Description", "Previous Offer", "Offer",
				}

				for idx, cr := range res {
					records[idx+1] = []string{
						cr.ActivityName, cr.Name, cr.Description,
						fmt.Sprintf("%d", cr.PreviousOffer),
						fmt.Sprintf("%d", cr.Offer),
					}
				}

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

			} else if jsonOutput {

				data, err := json.Marshal(res)
				if err != nil {
					fmt.Println(fmt.Errorf("Error on convert data to json: %s", err.Error()))
					os.Exit(1)
				}
				fmt.Println(string(data))
			} else {

				table := tablewriter.NewWriter(os.Stdout)
				table.SetBorders(tablewriter.Border{
					Left:   true,
					Top:    true,
					Right:  true,
					Bottom: true})
				table.SetHeader([]string{
					"Activity", "CR", "Description", "Previous Offer", "Offer",
				},
				)
				table.SetFooterAlignment(tablewriter.ALIGN_LEFT)
				table.SetColWidth(100)

				for _, cr := range res {
					table.Append([]string{
						cr.ActivityName, cr.Name, cr.Description,
						fmt.Sprintf("%d", cr.PreviousOffer),
						fmt.Sprintf("%d", cr.Offer),
					})
				}

				table.Render()
			}
		},
	}

	flags := cmd.Flags()
	flags.Bool("csv", false, "Print output in CSV format")
	flags.Bool("json", false, "Print output in JSON format")

	return cmd
}
