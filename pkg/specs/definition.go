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

type Client struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	File        string `json:"-" yaml:"-"`

	ActivitiesDirs []string `json:"activities_dirs,omitempty" yaml:"activities_dirs,omitempty"`

	Activities []Activity `json:"activities,omitempty" yaml:"activities,omitempty"`
}

type Scenario struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	File        string `json:"-" yaml:"-"`

	ResourceCosts []ResourceCost `json:"resources_cost,omitempty" yaml:"resources_cost,omitempty"`
	Rates         []ResourceRate `json:"rates,omitempty" yaml:"rates,omitempty"`

	NowTime string `json:"now,omitempty" yaml:"now,omitempty"`

	Scheduler string `json:"scheduler,omitempty" yaml:"scheduler,omitempty"`

	// For scheduler simple
	Tasks      []ScenarioTask     `json:"task_prorities,omitempty" yaml:"task_prorities,omitempty"`
	Activities []ScenarioActivity `json:"activities_priorities,omitempty" yaml:"activities_priorities,omitempty"`
}

type ScenarioTask struct {
	Name     string `json:"name" yaml:"name"`
	Priority int    `json:"priority" yaml:"priority"`

	OverrideResource []string `json:"override_resources,omitempty" yaml:"override_resources,omitempty"`
}

type ScenarioActivity struct {
	Name     string `json:"name" yaml:"name"`
	Priority int    `json:"priority" yaml:"priority"`
}

type ScenarioSchedule struct {
	*Scenario
	File string `json:"-" yaml:"-"`

	Schedule []TaskScheduled `json:schedule,omitempty yaml:"schedule,omitempty`
}

type Activity struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Note        string `json:"note,omitempty" yaml:"note,omitempty"`
	Priority    int    `json:"priority" yaml:"priority"`
	File        string `json:"-" yaml:"-"`
	Disabled    bool   `json:"disabled,omitempty" yaml:"disabled,omitempty"`
	Closed      bool   `json:"closed,omitempty" yaml:"closed,omitempty"`

	Offer        int64   `json:"offer,omitempty" yaml:"offer,omitempty"`
	TimeMaterial bool    `json:"time_material,omitempty" yaml:"time_material,omitempty"`
	TMDailyOffer float64 `json:"time_material_daily_offer,omitempty" yaml:"time_material_daily_offer,omitempty"`

	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Flags  []string          `json:"flags,omitempty" yaml:"flags,omitempty"`

	Tasks          []Task          `json:"tasks,omitempty" yaml:"tasks,omitempty"`
	ChangeRequests []ChangeRequest `json:"change_requests,omitempty" yaml:"change_requests,omitempty"`
}

type ChangeRequest struct {
	Name          string `json:"name" yaml:"name"`
	Description   string `json:"description,omitempty" yaml:"description,omitempty"`
	Note          string `json:"note,omitempty" yaml:"note,omitempty"`
	PreviousOffer int64  `json:"previous_offer,omitempty" yaml:"previous_offer,omitempty"`
	Offer         int64  `json:"offer,omitempty" yaml:"offer,omitempty"`

	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Flags  []string          `json:"flags,omitempty" yaml:"flags,omitempty"`
}

// General task structure for files specs
type Task struct {
	*Period

	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Note        string `json:"note,omitempty" yaml:"note,omitempty"`
	Priority    int    `json:"priority,omitempty" yaml:"priority,omitempty"`
	Effort      string `json:"effort,omitempty" yaml:"effort,omitempty"`
	Completed   bool   `json:"completed,omitempty" yaml:"completed,omitempty"`

	AllocatedResource []string `json:"resources,omitempty" yaml:"resources,omitempty"`

	Milestone string `json:"milestone,omitempty" yaml:"milestone,omitempty"`

	Flags  []string          `json:"flags,omitempty" yaml:"flags,omitempty"`
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`

	Tasks   []Task   `json:"subtasks,omitempty" yaml:"subtasks,omitempty"`
	Depends []string `json:"depends,omitempty" yaml:"depends,omitempty"`

	// Recursive options
	Recursive TaskRecursiveOpts `json:"recursive,omitempty" yaml:"recursive,omitempty"`
}

type TaskRecursiveOpts struct {
	Enable bool `json:"enable" yaml:"enable"`
	// Type of recursion: weekly (default) | monthly | daily
	Mode     string `json:"mode,omitempty" yaml:"mode,omitempty"`
	Duration string `json:"duration,omitempty" yaml:"duration,omitempty"`

	Exclude []Period `json:"exclude,omitempty" yaml:"exclude,omitempty"`
}

type Resource struct {
	User  string   `json:"user" yaml:"user"`
	Name  string   `json:"name" yaml:"name"`
	Email []string `json:"email,omitempty" yaml:"email,omitempty"`
	Phone []string `json:"phone,omitempty" yaml:"phone,omitempty"`
	File  string   `json:"-" yaml:"-"`

	Holidays   []ResourceHolidays   `json:"holidays,omitempty" yaml:"holidays,omitempty"`
	Sick       []ResourceSick       `json:"sick,omitempty" yaml:"sick,omitempty"`
	Unemployed []ResourceUnemployed `json:"unemployed,omitempty" yaml:"unemployed,omitempty"`
}

type Period struct {
	StartPeriod string `json:"start_period,omitempty" yaml:"start_period,omitempty"`
	EndPeriod   string `json:"end_period,omitempty" yaml:"end_period,omitempty"`

	StartTime int64 `json:"-" yaml:"-"`
	EndTime   int64 `json:"-" yaml:"-"`
}

type ResourceRate struct {
	*Period `json:"period" yaml:"period"`
	User    string  `json:"user" yaml:"user"`
	Rate    float64 `json:"rate" yaml:"rate"`
}

type ResourceCost struct {
	*Period `json:"period" yaml:"period"`
	User    string  `json:"user" yaml:"user"`
	Cost    float64 `json:"cost" yaml:"cost"`
}

type ResourceHolidays struct {
	*Period
}

type ResourceUnemployed struct {
	*Period
}

type ResourceSick struct {
	*Period
}

type AgendaTimesheets struct {
	File string `json:"-" yaml:"-"`
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	Timesheets []ResourceTimesheet `json:"timesheets" yaml:"timesheets"`
}

// ResourceTimesheet contains the description of the user
// timesheet for one single day. Multiple days range aren't supported.
type ResourceTimesheet struct {
	*Period
	User     string `json:"user" yaml:"user"`
	Task     string `json:"task" yaml:"task"`
	Duration string `json:"duration" yaml:"duration"`
	Note     string `json:"note,omitempty" yaml:"note,omitempty"`

	// Internal
	Cost    float64 `json:"cost,omitempty" yaml:"cost,omitempty"`
	Revenue float64 `json:"revenue,omitempty" yaml:"revenue,omitempty"`
}

type ResourceTsAggregated struct {
	*Period
	User     string `json:"user,omitempty" yaml:"user,omitempty"`
	Task     string `json:"task,omitempty" yaml:"task,omitempty"`
	Duration string `json:"duration" yaml:"duration"`
	Seconds  int64  `json:"-" yaml:"-"`

	Cost    float64 `json:"cost,omitempty" yaml:"cost,omitempty"`
	Revenue float64 `json:"revenue,omitempty" yaml:"revenue,omitempty"`
}

type TaskScheduled struct {
	*Task
	*Period

	// We don't need this on yaml because could be stored with clients
	Activity *Activity `json:"-" yaml:"-"`
	Client   *Client   `json:"client" yaml:"client"`

	Progress float64 `json:"progress,omitempty" yaml:"progress,omitempty"`
	WorkTime int64   `json:"work_time,omitempty" yaml:"work_time,omitempty"`
	LeftTime int64   `json:"-" yaml:"-"`

	Underestimated bool `json:"understimated,omitempty" yaml:"understimated,omitempty"`

	Timesheets []ResourceTimesheet `json:"timesheets,omitempty" yaml:"timesheets,omitempty"`
}
