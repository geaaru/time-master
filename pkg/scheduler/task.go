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

	specs "github.com/geaaru/time-master/pkg/specs"
	time "github.com/geaaru/time-master/pkg/time"
)

type RecursiveTaskSeer interface {
	DoPrevision(now string) error
	GetNextDay(date string) (string, error)
}

type DefaultRecursiveTaskSeer struct {
	Scheduler TimeMasterScheduler
	Task      *specs.TaskScheduled

	self RecursiveTaskSeer
}

type DailyRecursiveTaskSeer struct {
	*DefaultRecursiveTaskSeer
}

type WeeklyRecursiveTaskSeer struct {
	*DefaultRecursiveTaskSeer
}

type MonthlyRecursiveTaskSeer struct {
	*DefaultRecursiveTaskSeer
}

func NewRecursiveTaskSeer(s TimeMasterScheduler, t *specs.TaskScheduled) RecursiveTaskSeer {
	var ans RecursiveTaskSeer
	switch t.Task.Recursive.Mode {
	case "daily":
		ans = NewDailyRecursiveTaskSeer(s, t)
	case "monthly":
		ans = NewMonthlyRecursiveTaskSeer(s, t)
	default:
		ans = NewWeeklyRecursiveTaskSeer(s, t)
	}

	return ans
}

func newDefaultRecursiveTaskSeer(s TimeMasterScheduler, t *specs.TaskScheduled) *DefaultRecursiveTaskSeer {
	return &DefaultRecursiveTaskSeer{
		Scheduler: s,
		Task:      t,
	}
}

func NewDailyRecursiveTaskSeer(s TimeMasterScheduler, t *specs.TaskScheduled) *DailyRecursiveTaskSeer {
	ans := &DailyRecursiveTaskSeer{
		DefaultRecursiveTaskSeer: newDefaultRecursiveTaskSeer(s, t),
	}
	ans.self = ans
	return ans
}

func NewMonthlyRecursiveTaskSeer(s TimeMasterScheduler, t *specs.TaskScheduled) *MonthlyRecursiveTaskSeer {
	ans := &MonthlyRecursiveTaskSeer{
		DefaultRecursiveTaskSeer: newDefaultRecursiveTaskSeer(s, t),
	}
	ans.self = ans
	return ans
}

func NewWeeklyRecursiveTaskSeer(s TimeMasterScheduler, t *specs.TaskScheduled) *WeeklyRecursiveTaskSeer {
	ans := &WeeklyRecursiveTaskSeer{
		DefaultRecursiveTaskSeer: newDefaultRecursiveTaskSeer(s, t),
	}
	ans.self = ans
	return ans
}

func (r *DefaultRecursiveTaskSeer) DoPrevision(now string) error {
	var err error
	var workDate string

	notCompleted := true
	workDaySec, _ := time.ParseDuration("1d", r.Scheduler.GetConfig().GetWork().WorkHours)

	ok, err := time.IsAWorkDay(now)
	if err != nil {
		return err
	}
	if !ok {
		workDate, err = time.GetNextWorkDay(now)
		if err != nil {
			return err
		}
	} else {
		workDate = now
	}

	if r.Task.Task.Recursive.Duration == "" {
		return errors.New("Invalid recursive task " +
			r.Task.Task.Name + " without duration")
	}

	if r.Task.Task.Period.EndPeriod == "" {
		return errors.New("Invalid recursive task " +
			r.Task.Task.Name + " without end period")
	}

	// Get unix time of end period
	endTaskTime, err := time.ParseTimestamp(r.Task.Task.Period.EndPeriod, true)
	if err != nil {
		return err
	}

	for notCompleted {
		workTime, err := time.ParseTimestamp(workDate, true)
		if err != nil {
			return err
		}

		if workTime.Unix() > endTaskTime.Unix() {
			notCompleted = false
			break
		}

		// Check date is excluded
		available, err := r.Task.Task.Recursive.IsAvailable(workDate)
		if err != nil {
			return err
		}

		if !available {
			r.Scheduler.GetLogger().Debug(fmt.Sprintf(
				"[%s] [%s] Date excluded.", workDate, r.Task.Task.Name))
			// For monthly/weekly
			workDate, err = r.self.GetNextDay(workDate)
			if err != nil {
				return err
			}
			continue
		}

		nowTime, err := time.ParseTimestamp(workDate, true)
		if err != nil {
			return err
		}

		// Get total seconds for this time
		leftTime, err := time.ParseDuration(
			r.Task.Task.Recursive.Duration,
			r.Scheduler.GetConfig().GetWork().WorkHours,
		)
		if err != nil {
			return err
		}

		for leftTime > 0 {

			availableSecs := workDaySec
			userTime := int64(0)

			for _, resource := range r.Task.Task.AllocatedResource {

				r.Scheduler.GetLogger().Debug(fmt.Sprintf(
					"[%s] [%s] [%s] Allocate resource...", workDate, resource, r.Task.Task.Name))

				rdm, ok := (*r.Scheduler.GetResourcesMap())[resource]
				if !ok {
					return errors.New("Error on retrieve resource map for user " + resource)
				}

				// Check if the resource is available
				available, err := rdm.Resource.IsAvailable(workDate)
				if err != nil {
					return errors.New("Error on check resource availability for user " +
						resource + ": " + err.Error())
				}

				if !available {
					r.Scheduler.GetLogger().Debug(fmt.Sprintf(
						"[%s] [%s] [%s] Resource not available.", workDate, resource, r.Task.Task.Name))
					continue
				}

				if _, present := rdm.Days[workDate]; present {

					if rdm.Days[workDate] == 0 {
						r.Scheduler.GetLogger().Debug(fmt.Sprintf(
							"[%s] [%s] [%s] No more time for this day.", workDate,
							resource, r.Task.Task.Name))
						continue
					}

					availableSecs = rdm.Days[workDate]
					userTime = availableSecs

					r.Scheduler.GetLogger().Debug(fmt.Sprintf(
						"[%s] [%s] Available secs for user %d.", workDate, resource, availableSecs))

					if leftTime < availableSecs {
						userTime = leftTime
					}

				} else {
					// POST: no entry on resource daily map for this date

					userTime = workDaySec
					if leftTime < workDaySec {
						userTime = leftTime
					}

				}

				leftTime -= userTime
				r.Task.AddTimesheet(
					specs.NewResourceTimesheet(
						resource, workDate,
						r.Task.Task.Name,
						fmt.Sprintf("%ds", userTime),
					),
				)

				r.Scheduler.GetLogger().Debug(fmt.Sprintf(
					"[%s] [%s] [%s] Added %d sec. Left %d sec.",
					workDate, resource, r.Task.Task.Name, userTime, leftTime))

				rdm.Days[workDate] = availableSecs - userTime

				if leftTime == 0 {
					break
				}

				(*r.Scheduler.GetResourcesMap())[resource] = rdm

			}

			// For monthly/weekly
			workDate, err = r.self.GetNextDay(workDate)
			if err != nil {
				return err
			}

			wTime, err := time.ParseTimestamp(workDate, true)
			if err != nil {
				return err
			}

			if r.Task.Task.Recursive.Mode == "daily" && leftTime > 0 {
				return errors.New(
					"Too few resources for daily task " + r.Task.Task.Name)

			} else if r.Task.Task.Recursive.Mode == "weekly" && leftTime > 0 {
				_, weekNum1 := nowTime.ISOWeek()
				_, weekNum2 := wTime.ISOWeek()

				if weekNum1 != weekNum2 {
					return errors.New(
						fmt.Sprintf("Too few resources for weekly task %s for week %d",
							r.Task.Task.Name, weekNum1))
				}
			} else if r.Task.Task.Recursive.Mode == "monthly" && leftTime > 0 {
				if nowTime.Month() != wTime.Month() {
					return errors.New(
						"Too few resources for monthly task " + r.Task.Task.Name)
				}
			}

		} // end for leftime > 0

	}

	return nil
}

func (r *DailyRecursiveTaskSeer) GetNextDay(date string) (string, error) {
	return time.GetNextWorkDay(date)
}

func (r *WeeklyRecursiveTaskSeer) GetNextDay(date string) (string, error) {
	return time.GetNextWeekFirstWorkDay(date)
}

func (r *MonthlyRecursiveTaskSeer) GetNextDay(date string) (string, error) {
	return time.GetNextMonthFirstWorkDay(date)
}
