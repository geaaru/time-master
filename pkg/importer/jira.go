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
package importer

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	specs "github.com/geaaru/time-master/pkg/specs"

	"github.com/mudler/luet/pkg/helpers"
	"gopkg.in/yaml.v2"
)

type TmJiraImporter struct {
	*DefaultImporter
	ResourceMapping map[string]string
	IssueTaskMap    map[string]string
}

type TmJiraMapper struct {
	Resources []TmJiraResource `json:"resources" yaml:"resources"`
	Issues    []TmJiraIssue    `json:"issues" yaml:"issues"`
}

type TmJiraIssue struct {
	JiraIssue string `json:"jira_issue" yaml:"jira_issue"`
	TaskName  string `json:"task" yaml:"task"`
}

type TmJiraResource struct {
	JiraName string `json:"jira_name" yaml:"jira_name"`
	Name     string `json:"name" yaml:"name"`
}

type TmJiraCsvRow struct {
	Issue    string
	Descr    string
	Date     string
	WorkTime string
	User     string
}

func TmJiraMapperFromYaml(data []byte) (*TmJiraMapper, error) {
	ans := &TmJiraMapper{}
	if err := yaml.Unmarshal(data, ans); err != nil {
		return nil, err
	}
	return ans, nil
}

func NewTmJiraImporter(config *specs.TimeMasterConfig, tmDir, filePrefix string, opts ImportOpts) *TmJiraImporter {
	return &TmJiraImporter{
		DefaultImporter: NewDefaultImporter(config, tmDir, filePrefix, opts),
		ResourceMapping: make(map[string]string, 0),
		IssueTaskMap:    make(map[string]string, 0),
	}
}

func (i *TmJiraImporter) ImportMapper(mapper *TmJiraMapper) {
	if len(mapper.Resources) > 0 {
		for _, r := range mapper.Resources {
			i.ResourceMapping[r.JiraName] = r.Name
		}
	}

	if len(mapper.Issues) > 0 {
		for _, issue := range mapper.Issues {
			i.IssueTaskMap[issue.JiraIssue] = issue.TaskName
		}
	}
}

func (i *TmJiraImporter) LoadTimesheets(csvFile string) error {
	if !helpers.Exists(csvFile) {
		return errors.New("File " + csvFile + " not present")
	}

	data, err := ioutil.ReadFile(csvFile)
	if err != nil {
		return err
	}

	reader := csv.NewReader(strings.NewReader(string(data)))
	rowNum := 0
	jiraRows := []TmJiraCsvRow{}

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		rowNum++

		if rowNum == 1 {
			// Skip header
			continue
		}

		jiraRows = append(jiraRows, TmJiraCsvRow{
			Issue:    row[0],
			Descr:    strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(row[22], "\n", ""), "\r", "")),
			Date:     row[3],
			WorkTime: row[2],
			User:     row[5],
		})

		i.Logger.Debug("Parse row ", row)
	}

	return i.convertRows2Agenda(&jiraRows)
}

func (i *TmJiraImporter) convertRows2Agenda(rows *[]TmJiraCsvRow) error {

	if i.Opts.SplitResource {
		// Create a map where key is the user

		mAgenda := make(map[string]specs.AgendaTimesheets, 0)

		for _, row := range *rows {

			if userAgenda, ok := mAgenda[row.User]; ok {
				// POST: the user's agenda is been already created.
				rt := i.convertRow2ResourceTimesheet(&row)
				userAgenda.AddResourceTimesheet(rt)
				mAgenda[row.User] = userAgenda

			} else {
				// POST: The user's agenda is not present
				agenda := specs.AgendaTimesheets{}
				agenda.File = i.GetMappedUser(row.User)
				agenda.Name = agenda.File
				agenda.AddResourceTimesheet(i.convertRow2ResourceTimesheet(&row))
				mAgenda[row.User] = agenda

			}

		}

		// Add Agenda
		for _, agenda := range mAgenda {
			i.AddTimesheet(&agenda)
		}

	} else {
		// Create only one agenda

		agenda := specs.AgendaTimesheets{}
		for _, row := range *rows {
			agenda.AddResourceTimesheet(i.convertRow2ResourceTimesheet(&row))
		}

		i.AddTimesheet(&agenda)
	}

	return nil
}

func (i *TmJiraImporter) GetMappedUser(user string) (ans string) {
	if u, ok := i.ResourceMapping[user]; ok {
		ans = u
	} else {
		ans = user
	}
	return
}

func (i *TmJiraImporter) GetMappedTask(descr, issue string) (ans string) {
	// Check if there is an issue mapping
	if task, ok := i.IssueTaskMap[issue]; ok {
		ans = task
	} else {
		if descr != "" {
			ans = descr
		} else {
			ans = issue
		}
	}
	return
}

func (i *TmJiraImporter) convertRow2ResourceTimesheet(row *TmJiraCsvRow) *specs.ResourceTimesheet {
	ans := &specs.ResourceTimesheet{
		Period: &specs.Period{
			StartPeriod: row.Date,
		},
		User: i.GetMappedUser(row.User),
		Task: i.GetMappedTask(row.Descr, row.Issue),
		// To check. It seems that jira return time in hours.
		Duration: fmt.Sprintf("%sh", row.WorkTime),
	}

	return ans
}
