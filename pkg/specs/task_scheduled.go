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
	time "github.com/geaaru/time-master/pkg/time"
)

func (t *TaskScheduled) AddTimesheet(rt *ResourceTimesheet) {
	t.Timesheets = append(t.Timesheets, *rt)
}

func (t *TaskScheduled) ElaborateTimesheets(onlyDate bool, workHours int, withPlan bool) error {
	var err error
	minTime := int64(0)
	maxTime := int64(0)

	t.WorkTime = 0

	if len(t.Timesheets) > 0 {

		for _, rt := range t.Timesheets {

			wt, err := rt.GetSeconds(workHours)
			if err != nil {
				return err
			}
			t.WorkTime += wt

			d, err := rt.GetDateUnix(onlyDate)
			if err != nil {
				return err
			}

			if minTime == 0 || d < minTime {
				minTime = d
			}

			if maxTime == 0 || d > maxTime {
				maxTime = d
			}
		}
	}

	if minTime > 0 {
		t.Period.StartPeriod, err = time.Seconds2Date(minTime, onlyDate)
		if err != nil {
			return err
		}

		if t.Task.Completed || (withPlan && t.LeftTime == 0) {

			t.Period.EndPeriod, err = time.Seconds2Date(maxTime, onlyDate)
			if err != nil {
				return err
			}

			t.Period.EndTime = maxTime
		}
	}

	t.Period.StartTime = minTime

	return nil
}

type TaskSchedPrioritySorter []TaskScheduled

func (t TaskSchedPrioritySorter) Len() int      { return len(t) }
func (t TaskSchedPrioritySorter) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t TaskSchedPrioritySorter) Less(i, j int) bool {
	if t[i].Task.Priority == t[j].Task.Priority {
		return t[i].Task.Name > t[j].Task.Name
	}
	return t[i].Task.Priority < t[j].Task.Priority
}
