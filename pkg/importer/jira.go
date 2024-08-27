/*
Copyright (C) 2020-2024  Daniele Rondina <geaaru@gmail.com>
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
	"github.com/geaaru/time-master/pkg/tools"

	"gopkg.in/yaml.v2"
)

type TmJiraImporter struct {
	*DefaultImporter
	ResourceMapping map[string]string
	IssueTaskMap    map[string]string
	IgnoredIssueMap map[string]bool
	Before202009    bool
	Before202401    bool
	Before202402    bool
	Before202408    bool
}

type TmJiraMapper struct {
	Resources     []TmJiraResource `json:"resources" yaml:"resources"`
	Issues        []TmJiraIssue    `json:"issues" yaml:"issues"`
	IgnoredIssues []string         `json:"ignored_issues,omitempty" yaml:"ignored_issues,omitempty"`
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
		IgnoredIssueMap: make(map[string]bool, 0),
		Before202009:    false,
	}
}

func (i *TmJiraImporter) SetBefore202009() {
	i.Before202009 = true
}

func (i *TmJiraImporter) SetBefore202401() {
	i.Before202401 = true
}

func (i *TmJiraImporter) SetBefore202408() {
	i.Before202408 = true
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

	if len(mapper.IgnoredIssues) > 0 {
		for _, issue := range mapper.IgnoredIssues {
			i.IgnoredIssueMap[issue] = true
		}
	}
}

func (i *TmJiraImporter) LoadTimesheets(csvFile string) error {
	if !tools.Exists(csvFile) {
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

		dateIdx := 4
		descIdx := 33
		userIdx := 6
		if i.Before202009 {
			userIdx = 5
			dateIdx = 3
			descIdx = 22
		} else if i.Before202401 {
			userIdx = 5
			dateIdx = 3
			descIdx = 23
		} else if i.Before202402 {
			userIdx = 5
			dateIdx = 3
			descIdx = 30
		} else if i.Before202408 {
			userIdx = 5
			dateIdx = 3
			descIdx = 31
		}

		jiraRows = append(jiraRows, TmJiraCsvRow{
			Issue:    row[0],
			Descr:    strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(row[descIdx], "\n", ""), "\r", "")),
			Date:     row[dateIdx],
			WorkTime: row[2],
			User:     row[userIdx],
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

			if i.IsIssue2Ignore(row.Issue) {
				i.Logger.Debug("Ignoring issue " + row.Issue)
				continue
			}

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

func (i *TmJiraImporter) IsIssue2Ignore(issue string) bool {
	if ok := i.IgnoredIssueMap[issue]; ok {
		return true
	}
	return false
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
