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
	"errors"
	"fmt"
	"strings"
)

func (i *TimeMasterInstance) Validate(ignoreError bool) error {

	dupClients := 0
	dupActivities := 0
	dupTasks := 0
	clientsMap := make(map[string]bool, 0)

	activitiesMap := make(map[string]bool, 0)
	// Check tasks
	tasksMap := make(map[string]bool, 0)

	// Check for task or activities with dot on name
	for _, c := range i.Clients {

		if _, isPresent := clientsMap[c.Name]; isPresent {
			if !ignoreError {
				return errors.New("Duplicated client " + c.Name)
			}
			i.Logger.Warning("Found duplicated client " + c.Name)
			dupClients++
		} else {
			clientsMap[c.Name] = true
		}

		for _, a := range *c.GetActivities() {

			if _, isPresent := activitiesMap[a.Name]; isPresent {
				if !ignoreError {
					return errors.New("Duplicated activity " + a.Name)
				}

				i.Logger.Warning("Found duplicated activity " + a.Name + " for client " + c.Name)
				dupActivities++
			} else {
				activitiesMap[a.Name] = true
			}

			if strings.Contains(a.Name, ".") {
				i.Logger.Error(
					fmt.Sprintf("Activity name %s contains [.] that is a special char.",
						a.Name))
				if !ignoreError {
					return errors.New("Invalid name on activity " + a.Name)
				}
			}

			// Check name task
			for _, t := range *a.GetTasks() {
				err := t.Validate(ignoreError)
				if err != nil {
					if !ignoreError {
						return err
					}
					i.Logger.Warning("Invalid task " + t.Name + ": " + err.Error())
				}
			}

			for _, t := range a.GetAllTasksList() {

				i.Logger.Debug(
					fmt.Sprintf("[%s] [%s] Checking task %s...", c.Name, a.Name, t.Name))

				if _, isPresent := tasksMap[t.Name]; isPresent {
					if !ignoreError {
						return errors.New("Duplicated task " + t.Name + " on activity " + a.Name)
					}
					i.Logger.Warning("Found duplicated task " + t.Name + " on activity " + a.Name)
					dupTasks++
				} else {
					tasksMap[t.Name] = true
				}

			}

		}
	}

	// Check if timesheet are related to valid tasks
	for _, agenda := range i.Timesheets {

		for _, timesheet := range *agenda.GetTimesheets() {
			if ok := tasksMap[timesheet.Task]; !ok {

				errMsg := fmt.Sprintf("No task %s found for timesheet of user %s of period %s",
					timesheet.Task,
					timesheet.User,
					timesheet.Period.StartPeriod,
				)
				if !ignoreError {
					return errors.New(errMsg)
				}

				i.Logger.Warning(errMsg)
			}

		}
	}

	// Check depends of all tasks and resources allocations
	for _, c := range i.Clients {
		for _, a := range *c.GetActivities() {
			for _, t := range a.GetAllTasksList() {
				if len(t.Depends) > 0 {
					for _, depend := range t.Depends {
						if ok := tasksMap[depend]; !ok {
							errMsg := fmt.Sprintf("Invalid dependency %s for task %s",
								depend, t.Name)
							if !ignoreError {
								return errors.New(errMsg)
							}

							i.Logger.Warning(errMsg)
						}
					}
				}

				for _, t := range *a.GetTasks() {

					if len(t.AllocatedResource) == 0 && t.Milestone == "" {

						errMsg := fmt.Sprintf("No resources defined on task %s.%s",
							a.Name, t.Name)
						if !ignoreError {
							return errors.New(errMsg)
						}

						i.Logger.Warning(errMsg)
					}
				}
			}
		}
	}

	// Validate resources.
	for _, r := range i.Resources {
		err := r.Validate()
		if err != nil {
			errMsg := fmt.Sprintf("Invalid status of the resource %s: %s", r.User, err.Error())
			if !ignoreError {
				return errors.New(errMsg)
			}
			i.Logger.Warning(errMsg)
		}
	}

	return nil
}
