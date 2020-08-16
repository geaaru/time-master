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
	tablewriter "github.com/olekukonko/tablewriter"

	"github.com/spf13/cobra"
)

func NewListCommand(config *specs.TimeMasterConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "list of resources.",
		Run: func(cmd *cobra.Command, args []string) {

			// Create Instance
			tm := loader.NewTimeMasterInstance(config)

			err := tm.Load()
			if err != nil {
				fmt.Println("Error on load data:" + err.Error() + "\n")
				os.Exit(1)
			}

			resources := len(*tm.GetResources())
			table := tablewriter.NewWriter(os.Stdout)
			table.SetBorders(tablewriter.Border{
				Left:   true,
				Top:    true,
				Right:  true,
				Bottom: true})

			table.SetHeader([]string{"User", "Name"})
			table.SetFooterAlignment(tablewriter.ALIGN_RIGHT)

			userList := []string{}
			for _, c := range *tm.GetResources() {
				userList = append(userList, c.User)
			}
			sort.Strings(userList)

			for _, u := range userList {
				r := tm.GetResourceByUser(u)
				table.Append([]string{u, r.Name})
			}

			table.SetFooter([]string{
				"Total Resources: ",
				fmt.Sprintf("%d", resources),
			})

			table.Render()
		},
	}

	return cmd
}
