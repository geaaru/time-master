/*
Copyright (C) 2020-2025  Daniele Rondina <geaaru@macaronios.org>
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
	"time"

	specs "github.com/geaaru/time-master/pkg/specs"
	"github.com/geaaru/time-master/pkg/tools"

	"gopkg.in/yaml.v2"
)

type TmKimaiImporter struct {
	*DefaultImporter
	ResourceMapping  map[string]string
	ActivityTaskMap  map[string]string
	IgnoredLabelsMap map[string]bool
}

type TmKimaiMapper struct {
	Resources  []TmKimaiResource `json:"resources" yaml:"resources"`
	Activities []TmKimaiActivity `json:"activities" yaml:"activities"`

	IgnoredLabels []string `json:"ignored_labels,omitempty" yaml:"ignored_labels,omitempty"`
}

type TmKimaiResource struct {
	KimaiName string `json:"kimai_name" yaml:"jira_name"`
	Name      string `json:"name" yaml:"name"`
}

type TmKimaiActivity struct {
	Activity string `json:"activity" yaml:"activity"`
	Task     string `json:"task" yaml:"task"`
}

type TmKimaiCsvRow struct {
	Date     string
	WorkTime string
	User     string
	Project  string
	Activity string
	Descr    string
	Tags     []string
}

func TmKimaiMapperFromYaml(data []byte) (*TmKimaiMapper, error) {
	ans := &TmKimaiMapper{}
	if err := yaml.Unmarshal(data, ans); err != nil {
		return nil, err
	}
	return ans, nil
}

func NewTmKimaiImporter(config *specs.TimeMasterConfig, tmDir, filePrefix string, opts ImportOpts) *TmKimaiImporter {
	return &TmKimaiImporter{
		DefaultImporter:  NewDefaultImporter(config, tmDir, filePrefix, opts),
		ResourceMapping:  make(map[string]string, 0),
		ActivityTaskMap:  make(map[string]string, 0),
		IgnoredLabelsMap: make(map[string]bool, 0),
	}
}

func (i *TmKimaiImporter) ImportMapper(mapper *TmKimaiMapper) {
	if len(mapper.Resources) > 0 {
		for _, r := range mapper.Resources {
			i.ResourceMapping[r.KimaiName] = r.Name
		}
	}

	if len(mapper.Activities) > 0 {
		for _, activity := range mapper.Activities {
			i.ActivityTaskMap[activity.Activity] = activity.Task
		}
	}

	if len(mapper.IgnoredLabels) > 0 {
		for _, label := range mapper.IgnoredLabels {
			i.IgnoredLabelsMap[label] = true
		}
	}
}

func (i *TmKimaiImporter) LoadTimesheets(csvFile string) error {
	if !tools.Exists(csvFile) {
		return errors.New("File " + csvFile + " not present")
	}

	data, err := ioutil.ReadFile(csvFile)
	if err != nil {
		return err
	}

	reader := csv.NewReader(strings.NewReader(string(data)))
	rowNum := 0
	kimaiRows := []TmKimaiCsvRow{}

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

		kimaiRows = append(kimaiRows, TmKimaiCsvRow{
			// Date + From
			Date:     fmt.Sprintf("%s %s", row[0], row[1]),
			WorkTime: row[3],
			User:     row[10],
			Project:  row[13],
			Activity: row[14],
			Descr:    strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(row[15], "\n", ""), "\r", "")),
			Tags:     strings.Split(row[17], ","),
		})

		i.Logger.Debug("Parse row ", row)
	}

	return i.convertRows2Agenda(&kimaiRows)
}

func (i *TmKimaiImporter) convertRows2Agenda(rows *[]TmKimaiCsvRow) error {

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

func (i *TmKimaiImporter) IsLabel2Ignore(label string) bool {
	if ok := i.IgnoredLabelsMap[label]; ok {
		return true
	}
	return false
}

func (i *TmKimaiImporter) GetMappedUser(user string) (ans string) {
	if u, ok := i.ResourceMapping[user]; ok {
		ans = u
	} else {
		ans = user
	}
	return
}

func (i *TmKimaiImporter) GetMappedTask(activity string) (ans string) {
	// Check if there is an issue mapping
	if task, ok := i.ActivityTaskMap[activity]; ok {
		ans = task
	} else {
		ans = activity
	}
	return
}

func (i *TmKimaiImporter) convertRow2ResourceTimesheet(row *TmKimaiCsvRow) *specs.ResourceTimesheet {
	ans := &specs.ResourceTimesheet{
		Period: &specs.Period{
			StartPeriod: row.Date,
		},
		User: i.GetMappedUser(row.User),
	}

	// Check if there is mapping.
	task := i.GetMappedTask(strings.TrimSpace(row.Activity))

	if len(row.Tags) > 0 {
		// Get the first label not ignored.
		for idx := range row.Tags {
			if i.IsLabel2Ignore(strings.TrimSpace(row.Tags[idx])) {
				continue
			}
			label := strings.TrimSpace(row.Tags[idx])
			if label != "" {
				task = label
			}
		}
	}

	ans.Task = task

	// Convert duration HH:MM:SS in hours.
	parts := strings.Split(row.WorkTime, ":")
	if len(parts) != 3 {
		// TODO: manage better this.
		ans.Duration = "1h"
	} else {
		durationStr := fmt.Sprintf("%sh%sm%ss", parts[0], parts[1], parts[2])
		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			ans.Duration = "1h"
		} else {
			hours := duration.Hours()
			ans.Duration = fmt.Sprintf("%.1fh", hours)
		}
	}

	return ans
}
