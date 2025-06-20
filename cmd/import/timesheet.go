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
package cmd_client

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	importer "github.com/geaaru/time-master/pkg/importer"
	specs "github.com/geaaru/time-master/pkg/specs"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func loadMapperFile(file string) (*importer.TmJiraMapper, error) {
	fileAbs, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadFile(fileAbs)
	if err != nil {
		return nil, err
	}

	return importer.TmJiraMapperFromYaml(content)
}

func loadKimaiMapperFile(file string) (*importer.TmKimaiMapper, error) {
	fileAbs, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadFile(fileAbs)
	if err != nil {
		return nil, err
	}

	return importer.TmKimaiMapperFromYaml(content)
}

func NewTimesheetCommand(config *specs.TimeMasterConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "timesheet [file]",
		Short: "Import CSV timesheet",
		PreRun: func(cmd *cobra.Command, args []string) {

			if len(args) == 0 {
				fmt.Println("Missing import file!")
				os.Exit(1)
			}

			importType, _ := cmd.Flags().GetString("import-type")
			if importType != "jira" && importType != "kimai" {
				fmt.Println("import-type supported is only 'jira' or 'kimai'")
				os.Exit(1)
			}

			dir, _ := cmd.Flags().GetString("dir")
			if dir == "" {
				fmt.Println("Missing dir option")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			var imp importer.TimeMasterImporter

			importType, _ := cmd.Flags().GetString("import-type")
			dir, _ := cmd.Flags().GetString("dir")
			targetPrefix, _ := cmd.Flags().GetString("target-prefix")
			splitForUser, _ := cmd.Flags().GetBool("split-for-user")
			jiraMapperFile, _ := cmd.Flags().GetString("jira-mapper-file")
			kimaiMapperFile, _ := cmd.Flags().GetString("kimai-mapper-file")
			stdout, _ := cmd.Flags().GetBool("stdout")

			opts := importer.ImportOpts{
				SplitResource: splitForUser,
			}

			switch importType {
			case "jira":
				imp = importer.NewTmJiraImporter(config, dir, targetPrefix, opts)
			default:
				// Default kimai
				imp = importer.NewTmKimaiImporter(config, dir, targetPrefix, opts)
			}

			if jiraMapperFile != "" && importType == "jira" {
				mapper, err := loadMapperFile(jiraMapperFile)
				if err != nil {
					fmt.Println("Error on load file " + jiraMapperFile + ": " + err.Error())
					os.Exit(1)
				}
				(imp.(*importer.TmJiraImporter)).ImportMapper(mapper)

				jiraBefore202009, _ := cmd.Flags().GetBool("jira-before202009")
				jiraBefore202401, _ := cmd.Flags().GetBool("jira-before202401")
				jiraBefore202408, _ := cmd.Flags().GetBool("jira-before202408")
				if jiraBefore202009 {
					(imp.(*importer.TmJiraImporter)).SetBefore202009()
				}
				if jiraBefore202401 {
					(imp.(*importer.TmJiraImporter)).SetBefore202401()
				}
				if jiraBefore202408 {
					(imp.(*importer.TmJiraImporter)).SetBefore202408()
				}
			}

			if kimaiMapperFile != "" && importType == "kimai" {
				mapper, err := loadKimaiMapperFile(kimaiMapperFile)
				if err != nil {
					fmt.Println("Error on load file " + kimaiMapperFile + ": " + err.Error())
					os.Exit(1)
				}
				(imp.(*importer.TmKimaiImporter)).ImportMapper(mapper)
			}

			sourceFile := args[0]

			err := imp.LoadTimesheets(sourceFile)
			if err != nil {
				fmt.Println("Error on load file: " + err.Error())
				os.Exit(1)
			}

			if stdout {
				for _, agenda := range *imp.GetTimesheets() {
					data, err := yaml.Marshal(agenda)
					if err != nil {
						fmt.Println("Error on decode agenda in yaml: " + err.Error())
						os.Exit(1)
					}

					fmt.Println(string(data))
				}

			} else {

				err = imp.WriteTimesheets()
				if err != nil {
					fmt.Println("Error on write yaml file: " + err.Error())
					os.Exit(1)
				}

			}

		},
	}

	flags := cmd.Flags()
	flags.StringP("import-type", "i", "kimai",
		"Define type of the imported file. Supported values: jira|kimai.")
	flags.StringP("dir", "d", "", "Directory where import timesheets.")
	flags.StringP("target-prefix", "p", "", "Prefix of the file/files to create.")
	flags.BoolP("split-for-user", "s", false,
		"Create a timesheet file for every user.")
	flags.Bool("stdout", false, "Print timesheets to stdout instead of write files.")

	// Kimai options
	flags.StringP("kimai-mapper-file", "k", "", "Import Kimai resource mapper file.")

	// Jira options
	flags.StringP("jira-mapper-file", "j", "", "Import Jira resource mapper file.")
	flags.Bool("jira-before202009", false, "Import CSV created before 2020-09")
	flags.Bool("jira-before202401", false, "Import CSV created before 2024-01")
	flags.Bool("jira-before202408", false, "Import CSV created before 2024-08")

	return cmd
}
