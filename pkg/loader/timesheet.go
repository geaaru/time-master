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
package loader

import (
	"regexp"
	"strings"
	"time"

	specs "github.com/geaaru/time-master/pkg/specs"
	tmtime "github.com/geaaru/time-master/pkg/time"
)

func (i *TimeMasterInstance) CalculateTimesheetsCostAndRevenue(sName string) error {

	scenario, err := i.GetScenarioByName(sName)
	if err != nil {
		return err
	}

	for idx, _ := range i.Timesheets {
		agenda := &i.Timesheets[idx]

		for idx_t, rt := range agenda.Timesheets {

			cost, err := scenario.GetResourceCost4Date(rt.Period.StartPeriod, rt.User)
			if err != nil {
				return err
			}

			rate, err := scenario.GetResourceRate4Date(rt.Period.StartPeriod, rt.User)
			if err != nil {
				return err
			}

			workDaySec, _ := tmtime.ParseDuration("1d", i.Config.GetWork().WorkHours)
			secs, err := rt.GetSeconds(i.Config.GetWork().WorkHours)
			if err != nil {
				return err
			}

			activityName := rt.ResolveActivityByName()
			activity, _, err := i.GetActivityByName(activityName)
			if err != nil {
				return err
			}
			costRt := (cost / float64(workDaySec)) * float64(secs)

			var revenueRt float64
			if activity.IsTimeAndMaterial() {
				revenueRt = (activity.GetTimeAndMaterialDailyOffer() / float64(workDaySec)) * float64(secs)
			} else {
				revenueRt = (rate / float64(workDaySec)) * float64(secs)
			}

			agenda.Timesheets[idx_t].SetCost(costRt)
			agenda.Timesheets[idx_t].SetRevenue(revenueRt)

		}
	}

	return nil
}

func (i *TimeMasterInstance) GetAggregatedTimesheetsMap(opts specs.TimesheetResearch, from, to string, users []string, tasks []string) (map[string]*specs.ResourceTsAggregated, error) {

	var rta *specs.ResourceTsAggregated
	var fromDate, toDate time.Time
	var err error

	tsMap := make(map[string]*specs.ResourceTsAggregated)

	if from != "" {
		fromDate, err = tmtime.ParseTimestamp(from, true)
		if err != nil {
			return nil, err
		}
	}
	if to != "" {
		toDate, err = tmtime.ParseTimestamp(to, true)
		if err != nil {
			return nil, err
		}
	}

	for _, a := range *i.GetTimesheets() {

		for _, rt := range *a.GetTimesheets() {

			dUnix, _ := rt.GetDateUnix(true)
			if from != "" && dUnix < fromDate.Unix() {
				continue
			}
			if to != "" && dUnix > toDate.Unix() {
				continue
			}
			if len(users) > 0 && !matchEntry(rt.User, users) {
				continue
			}
			if len(tasks) > 0 && !regexEntry(rt.Task, tasks) {
				continue
			}

			key, err := rt.GetMapKey(opts, true)
			if err != nil {
				return nil, err
			}
			i.Logger.Debug("Using key ", key)
			if val, ok := tsMap[key]; ok {
				// POST: key already present
				rta = val
			} else {
				var task, user, date string

				if opts.Monthly {
					date, _ = rt.GetMonth(true)
				} else {
					date, _ = rt.GetDate(true)
				}

				if opts.ByActivity {
					leafs := strings.Split(rt.Task, ".")
					task = leafs[0]
				} else if opts.ByTask {
					task = rt.Task
				}
				if opts.ByUser {
					user = rt.User
				}
				rta = specs.NewResourceTsAggregated(date, user, task)
			}

			err = rta.AddResourceTimesheet(&rt, i.Config.GetWork().WorkHours)
			if err != nil {
				return nil, err
			}
			tsMap[key] = rta
		}

	}

	for _, v := range tsMap {
		v.CalculateDuration()
	}

	return tsMap, nil
}

func (i *TimeMasterInstance) GetAggregatedTimesheets(opts specs.TimesheetResearch, from, to string, users []string, tasks []string) (*[]specs.ResourceTsAggregated, error) {
	ans := []specs.ResourceTsAggregated{}

	tsMap, err := i.GetAggregatedTimesheetsMap(opts, from, to, users, tasks)
	if err != nil {
		return nil, err
	}

	for _, v := range tsMap {
		v.CalculateDuration()
		ans = append(ans, *v)
	}

	return &ans, nil
}

func matchEntry(entry string, whitelist []string) bool {
	for _, e := range whitelist {
		if entry == e {
			return true
		}
	}
	return false
}

func regexEntry(entry string, listRegex []string) bool {
	for _, e := range listRegex {
		r := regexp.MustCompile(e)
		if r != nil && r.MatchString(entry) {
			return true
		}
	}
	return false
}
