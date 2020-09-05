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
	"strconv"

	log "github.com/geaaru/time-master/pkg/logger"
	specs "github.com/geaaru/time-master/pkg/specs"
	time "github.com/geaaru/time-master/pkg/time"
)

type SimpleScheduler struct {
	*DefaultScheduler
}

func NewSimpleScheduler(config *specs.TimeMasterConfig, scenario *specs.Scenario) *SimpleScheduler {
	ans := &SimpleScheduler{
		DefaultScheduler: &DefaultScheduler{
			Config: config,
			Logger: log.NewTmLogger(config),
			Scenario: &specs.ScenarioSchedule{
				Scenario: scenario,
				Schedule: []specs.TaskScheduled{},
			},
			taskMap:      make(map[string]*specs.TaskScheduled, 0),
			ResourcesMap: make(map[string]*ResourceDailyMap, 0),
		},
	}

	// Initialize logging
	if config.GetLogging().EnableLogFile && config.GetLogging().Path != "" {
		err := ans.Logger.InitLogger2File()
		if err != nil {
			ans.Logger.Fatal("Error on initialize logfile")
		}
	}

	return ans
}

func (s *SimpleScheduler) BuildPrevision(opts SchedulerOpts) (*specs.ScenarioSchedule, error) {

	if opts.FilterPreElaboration {
		err := s.FilterPreElaboration(opts)
		if err != nil {
			return nil, err
		}
	}

	// Reset task map and schedule
	if len(s.Scenario.Schedule) > 0 {
		s.Scenario.Schedule = []specs.TaskScheduled{}
		s.taskMap = make(map[string]*specs.TaskScheduled, 0)
		s.ResourcesMap = make(map[string]*ResourceDailyMap, 0)
	}

	s.initializeTasks()

	// Assign resource timesheet to task scheduled
	err := s.assignTimesheets()
	if err != nil {
		return nil, err
	}

	// Elaborate task scheduled
	err = s.elaborateTimesheets(false)
	if err != nil {
		return nil, err
	}

	if !opts.SkipPlan {
		err = s.doPrevision()
		if err != nil {
			return nil, err
		}

		// Elaborate task scheduled
		err = s.elaborateTimesheets(true)
		if err != nil {
			return nil, err
		}
	}

	err = s.FilterPostElaboration(opts)
	if err != nil {
		return nil, err
	}

	return s.Scenario, nil
}

func (s *SimpleScheduler) doPrevision() error {
	var err error

	tasks := []specs.TaskScheduled{}
	completedTasks := []specs.TaskScheduled{}

	s.initResourceMap()

	// Retrieve the list of not closed tasks with effort.
	for idx, ts := range s.Scenario.Schedule {
		if ts.Task.Completed || ts.Task.Effort == "" {
			continue
		}

		// Calculate effort in seconds
		effortSecs, err := ts.Task.GetEffortSeconds(s.Config.GetWork().WorkHours)
		if err != nil {
			return err
		}

		if ts.WorkTime > effortSecs {
			s.Scenario.Schedule[idx].Underestimated = true
			s.Scenario.Schedule[idx].LeftTime = 0
			// TODO: check if add on array or not
			continue
		} else if ts.WorkTime == effortSecs {
			// POST: I consider closed the task
			s.Scenario.Schedule[idx].LeftTime = 0
			continue
		}

		// Calculate progress before assign all resource timesheet
		s.Scenario.Schedule[idx].Progress, _ = strconv.ParseFloat(
			fmt.Sprintf("%02.02f", (float64(ts.WorkTime)/float64(effortSecs))*100), 64)

		s.Scenario.Schedule[idx].LeftTime = effortSecs - ts.WorkTime

		tasks = append(tasks, s.Scenario.Schedule[idx])
	}

	// Sort task for priority
	sort.Sort(specs.TaskSchedPrioritySorter(tasks))

	workDate := s.Scenario.NowTime
	workDaySec, _ := time.ParseDuration("1d", s.Config.GetWork().WorkHours)
	for len(tasks) > 0 {
		workDate, err = time.GetNextWorkDay(workDate)
		if err != nil {
			return err
		}

		inProgressTasks := []specs.TaskScheduled{}

		for idx, t := range tasks {

			completed := false

			if len(t.AllocatedResource) == 0 {
				return errors.New(fmt.Sprintf("No resources for task %s", t.Name))
			}

			availableSecs := workDaySec
			workTime := int64(0)

			for _, r := range t.AllocatedResource {

				s.Logger.Debug(fmt.Sprintf(
					"[%s] [%s] [%s] Allocate resource...", workDate, r, t.Name))

				rdm, ok := s.ResourcesMap[r]
				if !ok {
					return errors.New("Error on retrieve resource map for user " + r)
				}

				if _, present := rdm.Days[workDate]; present {

					if rdm.Days[workDate] == 0 {
						s.Logger.Debug(fmt.Sprintf(
							"[%s] [%s] [%s] No more time for this day.", workDate, r, t.Name))
						continue
					}

					availableSecs = rdm.Days[workDate]
					workTime = availableSecs

					s.Logger.Debug(fmt.Sprintf(
						"[%s] [%s] Available secs for user %d.", workDate, r, availableSecs))

					if tasks[idx].LeftTime < availableSecs {
						workTime = tasks[idx].LeftTime
					}

				} else {
					// POST: no entry on resource daily map for this date

					workTime = workDaySec
					if tasks[idx].LeftTime < workDaySec {
						workTime = tasks[idx].LeftTime
					}

				}

				tasks[idx].LeftTime -= workTime
				tasks[idx].AddTimesheet(
					specs.NewResourceTimesheet(
						r, workDate,
						tasks[idx].Task.Name,
						fmt.Sprintf("%ds", workTime),
					),
				)

				s.Logger.Debug(fmt.Sprintf(
					"[%s] [%s] [%s] Added %d sec. Left %d sec.",
					workDate, r, t.Name, workTime, tasks[idx].LeftTime))

				rdm.Days[workDate] = availableSecs - workTime

				if tasks[idx].LeftTime == 0 {
					completed = true
					completedTasks = append(completedTasks, tasks[idx])
					break
				}

				s.ResourcesMap[r] = rdm

			}

			if !completed {
				inProgressTasks = append(inProgressTasks, tasks[idx])
			}
		}

		tasks = inProgressTasks
	}

	// POST: all tasks are been completed
	for _, t := range completedTasks {
		s.taskMap[t.Name].Timesheets = t.Timesheets
		s.taskMap[t.Name].LeftTime = 0
	}

	return nil
}
