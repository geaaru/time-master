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

type ActivityReport struct {
	*Activity

	RevenuePlan float64 `json:"revenue_plan,omitempty" yaml:"revenue_plan,omitempty"`
	Cost        float64 `json:"cost,omitempty" yaml:"cost,omitempty"`
	Profit      float64 `json:"profit,omitempty" yaml:"profit,omitempty"`

	Work       string `json:"work,omitempty" yaml:"work,omitempty"`
	WorkPerc   string `json:"work_perc,omitempty" yaml:"work_perc,omitempty"`
	Duration   string `json:"duration,omitempty" yaml:"duration,omitempty"`
	ProfitPerc string `json:"profit_perc,omitempty" yaml:"profit_perc,omitempty"`

	BusinessProgressPerc string `json:"business_progress_perc,omitempty" yaml:"business_progress_perc,omitempty"`

	Effort   int64 `json:"effort_sec,omitempty" yaml:"effort_sec,omitempty"`
	WorkSecs int64 `json:"work_sec,omitempty" yaml:"work_sec,omitempty"`
}

type TimesheetReport struct {
	User      string `json:"user,omitempty" yaml:"user,omitempty"`
	Date      string `json:"date,omitempty" yaml:"date,omitempty"`
	Task      string `json:"task,omitempty" yaml:"task,omitempty"`
	Activity  string `json:"activity,omitempty" yaml:"activity,omitempty"`
	Effort    string `json:"effort,omitempty" yaml:"effort,omitempty"`
	EffortSec int64  `json:"effort_sec,omitempty" yaml:"effort_sec,omitempty"`
}

type TimesheetReportPerUser struct {
	User         string                 `json:"user,omitempty" yaml:"user,omitempty"`
	Events       []TimesheetReportEvent `json:"events,omitempty", yaml:"events,omitempty"`
	TotEffortSec int64                  `json:"tot_effort_sec,omitempty" yaml:"tot_effort_sec,omitempty"`
	TotEffort    string                 `json:"tot_effort,omitempty" yaml:"tot_effort,omitempty"`
}

type TimesheetReportEvent struct {
	Date      string `json:"date" yaml:"date"`
	Effort    string `json:"effort" yaml:"effort"`
	EffortSec int64  `json:"effort_sec" yaml:"effort_sec"`
}

type TaskReport struct {
	*Task `json:"task,omitempty" yaml:"task,omitempty"`

	Name        string  `json:"name" yaml:"name"`
	Description string  `json:"description,omitempty" yaml:"description,omitempty"`
	Work        string  `json:"work,omitempty" yaml:"work,omitempty"`
	WorkSec     int64   `json:"work_sec,omitempty" yaml:"work_sec,omitempty"`
	Effort      string  `json:"effort" yaml:"effort"`
	EffortSec   int64   `json:"effort_sec" yaml:"effort_sec"`
	Cost        float64 `json:"cost,omitempty" yaml:"cost,omitempty"`
}

type ChangeRequestReport struct {
	*ChangeRequest
	ActivityName string `json:"activity,omitempty" yaml:"activity,omitempty"`
}
