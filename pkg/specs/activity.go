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
	"errors"
	"strings"

	"gopkg.in/yaml.v2"
)

func ActivityFromYaml(data []byte, file string) (*Activity, error) {
	ans := &Activity{}
	if err := yaml.Unmarshal(data, ans); err != nil {
		return nil, err
	}
	ans.File = file

	return ans, nil
}

func NewActivity(name, description string) *Activity {
	return &Activity{
		Name:        name,
		Description: description,
		Note:        "",
		Priority:    100,
		Disabled:    false,
		Closed:      false,
		Labels:      make(map[string]string, 0),
		Flags:       []string{},
		Tasks:       []Task{},
	}
}

func (a *Activity) AddTask(t *Task) {
	a.Tasks = append(a.Tasks, *t)
}

func (a *Activity) IsClosed() bool                        { return a.Closed }
func (a *Activity) GetOffer() int64                       { return a.Offer }
func (a *Activity) IsTimeAndMaterial() bool               { return a.TimeMaterial }
func (a *Activity) GetTimeAndMaterialDailyOffer() float64 { return a.TMDailyOffer }

func (a *Activity) GetPlannedEffortTotSecs(workHours int) (int64, error) {
	var ans int64

	ans = 0
	if len(a.Tasks) > 0 {
		for _, t := range a.Tasks {
			e, err := t.GetPlannedEffortTotSecs(workHours)
			if err != nil {
				return -1, err
			}

			ans += e
		}
	}

	return ans, nil
}

func (a *Activity) GetTasks() *[]Task {
	return &a.Tasks
}

func (a *Activity) GetTaskByFullName(name string) (*Task, error) {
	var ans *Task = nil

	leafs := strings.Split(name, ".")
	if len(leafs) == 1 {
		return nil, errors.New("Invalid name " + name + " without .")
	}

	if leafs[0] != a.Name {
		return nil, errors.New("Invalid task name " + name + " without prefix " + a.Name)
	}

	mainActivity := leafs[1]

	for idx, t := range a.Tasks {
		if t.Name == mainActivity {
			ans = &a.Tasks[idx]
			break
		}
	}

	if ans == nil {
		return nil, errors.New("No task found with leaf " + mainActivity)
	}

	if len(leafs) > 2 {
		return ans.GetTaskByFullName(strings.Join(leafs[1:], "."))
	}

	return ans, nil
}

func (a *Activity) GetAllTasksList() []Task {
	ans := []Task{}

	if len(a.Tasks) > 0 {
		for _, t := range a.Tasks {
			ans = append(ans, t.GetAllTasksAndSubTasksList(a.Name, []string{})...)
		}
	}

	// Ensure that task are all completed for closed activity
	if a.Closed {
		for idx, _ := range ans {
			ans[idx].Completed = true
		}
	}

	return ans
}

func (a *Activity) HasLabelKey(key string) bool {
	for k, _ := range a.Labels {
		if k == key {
			return true
		}
	}
	return false
}

func (a *Activity) InitDefaultPriority(prio int) {
	if a.Priority == 0 {
		a.Priority = prio
	}
	if len(a.Tasks) > 0 {
		for idx, _ := range a.Tasks {
			a.Tasks[idx].InitDefaultPriority(prio)
		}
	}
}
