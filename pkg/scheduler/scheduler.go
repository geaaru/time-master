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
package scheduler

import (
	"errors"
	"fmt"
	"sort"

	log "github.com/geaaru/time-master/pkg/logger"
	specs "github.com/geaaru/time-master/pkg/specs"
	time "github.com/geaaru/time-master/pkg/time"
)

type TimeMasterScheduler interface {
	BuildPrevision(SchedulerOpts) (*specs.ScenarioSchedule, error)
	SetClients(*[]specs.Client)
	SetResources(*[]specs.Resource)
	SetTimesheets(*[]specs.AgendaTimesheets)
}

type SchedulerOpts struct {
	Clients              []string
	Activities           []string
	ExcludeTaskFlags     []string
	FilterPreElaboration bool
	OnlyClosed           bool
	SkipEmptyTasks       bool
}

type DefaultScheduler struct {
	Logger     *log.TmLogger
	Config     *specs.TimeMasterConfig
	Clients    []specs.Client
	Resources  []specs.Resource
	Timesheets []specs.AgendaTimesheets
	Scenario   *specs.ScenarioSchedule
	taskMap    map[string]*specs.TaskScheduled
}

func (s *DefaultScheduler) FilterPostElaboration(opts SchedulerOpts) error {

	if !opts.FilterPreElaboration {

		if len(opts.Clients) > 0 {

			tasks := []specs.TaskScheduled{}

			for idx, ts := range s.Scenario.Schedule {

				found := false
				for _, c := range opts.Clients {
					if ts.Client.GetName() == c {
						found = true
						break
					}
				}

				if found {
					tasks = append(tasks, s.Scenario.Schedule[idx])
				}

			}

			s.Scenario.Schedule = tasks

		}

		if len(opts.Activities) > 0 {

			tasks := []specs.TaskScheduled{}

			for idx, ts := range s.Scenario.Schedule {

				found := false
				for _, c := range opts.Activities {
					if ts.Activity.Name == c {
						found = true
						break
					}
				}

				if found {
					tasks = append(tasks, s.Scenario.Schedule[idx])
				}

			}

			s.Scenario.Schedule = tasks
		}
	}

	if opts.OnlyClosed {
		tasks := []specs.TaskScheduled{}

		for idx, ts := range s.Scenario.Schedule {
			if ts.Task.Completed {
				if opts.SkipEmptyTasks && ts.Period.StartTime == 0 {
					continue
				}

				tasks = append(tasks, s.Scenario.Schedule[idx])
			}
		}
		s.Scenario.Schedule = tasks
	}

	if len(opts.ExcludeTaskFlags) > 0 {
		tasks := []specs.TaskScheduled{}

		for idx, ts := range s.Scenario.Schedule {

			excluded := false
			for _, exclude := range opts.ExcludeTaskFlags {

				if ts.Task.HasFlag(exclude) {
					excluded = true
					break
				}
			}

			if !excluded {
				tasks = append(tasks, s.Scenario.Schedule[idx])
			}
		}
		s.Scenario.Schedule = tasks
	}

	return nil
}

func (s *DefaultScheduler) FilterPreElaboration(opts SchedulerOpts) error {

	if !opts.FilterPreElaboration {
		return errors.New("Unexpected filter pre elaboration called.")
	}

	if len(opts.Clients) > 0 {
		clients := []specs.Client{}

		for _, c := range opts.Clients {
			for _, client := range s.Clients {
				if client.Name == c {
					clients = append(clients, client)
					break
				}
			}
		}

		if len(clients) == 0 {
			return errors.New("No clients available after the filter")
		}
		s.Clients = clients
	}

	if len(opts.Activities) > 0 {

		for idx, client := range s.Clients {

			activities := []specs.Activity{}

			for _, a := range opts.Activities {
				for _, activity := range *client.GetActivities() {
					if activity.Name == a {
						activities = append(activities, activity)
						break
					}
				}
			}

			s.Clients[idx].Activities = activities
		}
	}

	return nil
}

func (s *DefaultScheduler) SetClients(clients *[]specs.Client) {
	s.Clients = *clients
}

func (s *DefaultScheduler) SetResources(r *[]specs.Resource) {
	s.Resources = *r
}

func (s *DefaultScheduler) SetTimesheets(t *[]specs.AgendaTimesheets) {
	s.Timesheets = *t
}

func (s *DefaultScheduler) initializeTasks() {

	// Retrieve the list of all tasks
	for idx_c, client := range s.Clients {

		for _, activity := range *client.GetActivities() {
			aTasks := activity.GetAllTasksList()
			s.Scenario.Schedule = append(s.Scenario.Schedule, s.convertTasks2TaskScheduled(
				&aTasks,
				&s.Clients[idx_c],
				activity,
			)...)
		}

	}

	// Create a map of tasks
	for idx, t := range s.Scenario.Schedule {
		s.taskMap[t.Task.Name] = &s.Scenario.Schedule[idx]
		if t.Task.Completed {
			s.Scenario.Schedule[idx].Progress = 100.0
		}
		if t.Task.Priority == 0 {
			s.taskMap[t.Task.Name].Task.Priority = s.Config.GetWork().TaskDefaultPriority
		}
	}

	// Apply activity priorities
	if len(s.Scenario.Scenario.Activities) > 0 {
		for idx, ts := range s.Scenario.Schedule {

			for _, sa := range s.Scenario.Scenario.Activities {

				if ts.Activity.Name == sa.Name {
					s.Scenario.Schedule[idx].Task.Priority = sa.Priority
					break
				}
			}

		}
	}

	if len(s.Scenario.Scenario.Tasks) > 0 {

		for idx, ts := range s.Scenario.Schedule {
			for _, st := range s.Scenario.Scenario.Tasks {
				if ts.Task.Name == st.Name {
					s.Scenario.Schedule[idx].Task.Priority = st.Priority

					if len(st.OverrideResource) > 0 {
						s.Scenario.Schedule[idx].Task.AllocatedResource = st.OverrideResource
					}

					break
				}
			}
		}
	}

}

func (s *DefaultScheduler) elaborateTimesheets(withPlan bool) error {

	// 1. Elaborate timesheet and calculate start / end of the task with effort.
	for _, st := range s.taskMap {
		// TODO: Handle only date correctly
		err := st.ElaborateTimesheets(true, s.Config.GetWork().WorkHours, withPlan)
		if err != nil {
			return err
		}
	}

	// 2.Elaborate father tasks and milestone of completed tasks
	err := s.elaborateFatherTasksAndMilestone(withPlan)
	if err != nil {
		return err
	}

	return nil
}

func (s *DefaultScheduler) elaborateFatherTasksAndMilestone(withPlan bool) error {
	var err error

	// Create a list of milestone tasks for final elaboration
	mTasks := []*specs.TaskScheduled{}

	// Create list of task without timesheets but closed
	closedTasks := []*specs.TaskScheduled{}

	mKeys := []string{}
	for k, _ := range s.taskMap {
		mKeys = append(mKeys, k)
	}

	// Sort key on reverse order. In this way the father of the father is
	// handled in the right order.
	sort.Sort(sort.Reverse(sort.StringSlice(mKeys)))

	for _, k := range mKeys {

		st, ok := s.taskMap[k]
		if !ok {
			return errors.New("Error on retrieve task " + k + " from map")
		}

		if st.Task.Milestone != "" {
			mTasks = append(mTasks, st)
			continue
		}

		if st.Task.Effort != "" && len(st.Task.Tasks) == 0 && len(st.Timesheets) == 0 && st.Task.Completed {

			closedTasks = append(closedTasks, st)
			continue
		}

		if st.Task.Effort == "" {

			if len(st.Task.Tasks) > 0 {

				minTime := int64(0)
				maxTime := int64(0)

				for _, task := range st.Task.Tasks {

					// Retrieve task scheduled of the childer
					cst, ok := s.taskMap[st.Task.Name+"."+task.Name]
					if !ok {
						return errors.New(fmt.Sprintf(
							"Error on retrieve task %s from map of the father %s.",
							task.Name, st.Task.Name))
					}

					if minTime == 0 || (cst.Period.StartTime < minTime && cst.Period.StartTime > 0) {
						minTime = cst.Period.StartTime
					}

					if st.Task.Completed || (withPlan && st.LeftTime == 0) {
						if maxTime == 0 || cst.Period.EndTime > maxTime {
							maxTime = cst.Period.EndTime
						}
					}
				}

				st.Period.StartTime = minTime
				st.Period.EndTime = maxTime

				if minTime > 0 {
					// TODO: handle onlyDate correctly
					st.Period.StartPeriod, err = time.Seconds2Date(minTime, true)
					if err != nil {
						return err
					}
				}

				if maxTime > 0 && (st.Task.Completed || withPlan && st.LeftTime == 0) {
					st.Period.EndPeriod, err = time.Seconds2Date(maxTime, true)
					if err != nil {
						return err
					}
				}

			}
		}

	}

	// Fix closed tasks without timesheet
	for _, st := range closedTasks {
		if len(st.Task.Depends) > 0 {

			minTime := int64(0)
			maxTime := int64(0)

			st = s.taskMap[st.Task.Name]

			for _, task := range st.Task.Depends {

				// Retrieve task scheduled of the childer
				cst, ok := s.taskMap[task]
				if !ok {
					return errors.New(fmt.Sprintf(
						"Error on retrieve task %s from map of the father %s.",
						task, st.Task.Name))
				}

				if minTime == 0 || (cst.Period.StartTime < minTime && cst.Period.StartTime > 0) {
					minTime = cst.Period.StartTime
				}

				if st.Task.Completed || (withPlan && st.LeftTime == 0) {
					if maxTime == 0 || cst.Period.EndTime > maxTime {
						maxTime = cst.Period.EndTime
					}
				}
			}

			st.Period.StartTime = minTime
			st.Period.EndTime = maxTime

			if minTime > 0 {
				// TODO: handle onlyDate correctly
				st.Period.StartPeriod, err = time.Seconds2Date(minTime, true)
				if err != nil {
					return err
				}
			}

			if maxTime > 0 && (st.Task.Completed || (withPlan && st.LeftTime == 0)) {
				st.Period.EndPeriod, err = time.Seconds2Date(maxTime, true)
				if err != nil {
					return err
				}
			}

		}
	}

	// Fix milestone
	for _, st := range mTasks {

		s.Logger.Debug("Update period of the milestone " + st.Task.Name)

		minTime := int64(0)
		maxTime := int64(0)

		st = s.taskMap[st.Task.Name]

		for _, task := range st.Task.Depends {

			// Retrieve task scheduled of the childer
			cst, ok := s.taskMap[task]
			if !ok {
				return errors.New(fmt.Sprintf(
					"Error on retrieve task %s from map of the father %s.",
					task, st.Task.Name))
			}

			if minTime == 0 || (cst.Period.StartTime < minTime && cst.Period.StartTime > 0) {
				minTime = cst.Period.StartTime
			}

			if st.Task.Completed || (withPlan && st.LeftTime == 0) {
				if maxTime == 0 || cst.Period.EndTime > maxTime {
					maxTime = cst.Period.EndTime
				}
			}
		}

		st.Period.StartTime = minTime
		st.Period.EndTime = maxTime

		if minTime > 0 {
			// TODO: handle onlyDate correctly
			st.Period.StartPeriod, err = time.Seconds2Date(minTime, true)
			if err != nil {
				return err
			}
		}

		if maxTime > 0 && (st.Task.Completed || withPlan && st.LeftTime == 0) {
			st.Period.EndPeriod, err = time.Seconds2Date(maxTime, true)
			if err != nil {
				return err
			}

			s.Logger.Debug(fmt.Sprintf("Milestone %s completed from %s to %s.",
				st.Task.Name, st.Period.StartPeriod, st.Period.EndPeriod))

		}

	}

	return nil
}

func (s *DefaultScheduler) assignTimesheets() error {

	for _, agenda := range s.Timesheets {

		for _, timesheet := range *agenda.GetTimesheets() {

			if _, ok := s.taskMap[timesheet.Task]; ok {
				s.taskMap[timesheet.Task].AddTimesheet(&timesheet)
			} else {
				s.Logger.Debug("Task " + timesheet.Task + " not found for assign timesheet")
			}

		}
	}

	return nil
}

func (s *DefaultScheduler) convertTasks2TaskScheduled(tasks *[]specs.Task, c *specs.Client, a specs.Activity) []specs.TaskScheduled {

	ans := []specs.TaskScheduled{}

	for _, t := range *tasks {

		// Clone the object to avoid reuse of the pointer
		var task specs.Task = t
		ans = append(ans, specs.TaskScheduled{
			Task:     &task,
			Activity: &a,
			Client:   c,
			Period: &specs.Period{
				StartPeriod: "",
				EndPeriod:   "",
			},
		})

	}

	return ans
}