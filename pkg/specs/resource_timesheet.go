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
package specs

import (
	"fmt"
	"strings"

	"github.com/geaaru/time-master/pkg/time"
)

func (rt *ResourceTimesheet) GetDate(onlyDate bool) (string, error) {
	date, err := time.ParseTimestamp(rt.Period.StartPeriod, onlyDate)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d-%02d-%02d", date.Year(), date.Month(), date.Day()), nil
}

func (rt *ResourceTimesheet) GetDateUnix(onlyDate bool) (int64, error) {
	date, err := time.ParseTimestamp(rt.Period.StartPeriod, onlyDate)
	if err != nil {
		return 0, err
	}
	return date.Unix(), nil
}

func (rt *ResourceTimesheet) GetMonth(onlyDate bool) (string, error) {
	date, err := time.ParseTimestamp(rt.Period.StartPeriod, onlyDate)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d-%02d", date.Year(), date.Month()), nil
}

func (rt *ResourceTimesheet) GetMapKey(opts TimesheetResearch, onlyDate bool) (string, error) {
	var ans string = ""

	date, err := time.ParseTimestamp(rt.Period.StartPeriod, onlyDate)
	if err != nil {
		return "", err
	}

	if opts.ByUser {
		ans += fmt.Sprintf("%s-", rt.User)
	}

	if opts.ByActivity {
		leafs := strings.Split(rt.Task, ".")
		ans += fmt.Sprintf("%s-", leafs[0])
	} else if opts.ByTask {
		ans += fmt.Sprintf("%s-", rt.Task)
	}

	if !opts.IgnoreTime {
		if opts.Monthly {
			ans += fmt.Sprintf("%d-%02d", date.Year(), date.Month())
		} else {
			ans += fmt.Sprintf("%d-%02d-%02d", date.Year(), date.Month(), date.Day())
		}
	}

	if strings.HasSuffix(ans, "-") {
		ans = ans[:len(ans)-1]
	}

	return ans, nil
}

func NewResourceTsAggregated(date, user, task string) *ResourceTsAggregated {
	return &ResourceTsAggregated{
		Period: &Period{
			StartPeriod: date,
		},
		User:    user,
		Task:    task,
		Seconds: 0,
	}
}

func (rta *ResourceTsAggregated) AddResourceTimesheet(rt *ResourceTimesheet, workHours int) error {
	secs, err := time.ParseDuration(rt.Duration, workHours)
	if err != nil {
		return err
	}
	rta.Seconds += secs

	return nil
}

func (rta *ResourceTsAggregated) CalculateDuration() (ans error) {
	rta.Duration, ans = time.Seconds2Duration(rta.Seconds)
	return
}

func (rta *ResourceTsAggregated) GetDuration() string {
	return rta.Duration
}

func (rta *ResourceTsAggregated) GetSeconds() int64 {
	return rta.Seconds
}
