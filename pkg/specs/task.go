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
	"fmt"
	"strings"

	time "github.com/geaaru/time-master/pkg/time"
)

func (t *Task) GetPlannedEffortTotSecs(workHours int) (int64, error) {
	var ans int64
	var err error

	if t.Effort != "" {
		ans, err = time.ParseDuration(t.Effort, workHours)
		if err != nil {
			return -1, err
		}
	}

	if len(t.Tasks) > 0 {
		for _, subtask := range t.Tasks {
			e, err := subtask.GetPlannedEffortTotSecs(workHours)
			if err != nil {
				return -1, err
			}

			ans += e
		}
	}

	return ans, nil
}

func (t *Task) GetEffort() string {
	return t.Effort
}

func (t *Task) GetEffortSeconds(workHours int) (int64, error) {
	var ans int64 = 0
	var err error

	if t.Effort != "" {
		ans, err = time.ParseDuration(t.Effort, workHours)
		if err != nil {
			return -1, err
		}
	}

	return ans, err
}

func (t *Task) IsCompleted() bool {
	return t.Completed
}

func (t *Task) GetAllTasksAndSubTasksList(fatherName string, fatherResources []string) []Task {
	var fullName string
	ans := []Task{*t}

	if fatherName != "" {
		fullName = fmt.Sprintf("%s.%s", fatherName, t.Name)
	} else {
		fullName = t.Name
	}

	ans[0].Name = fullName

	// If task is without resources I allocate father
	// resources
	if len(t.AllocatedResource) == 0 && len(fatherResources) > 0 {
		ans[0].AllocatedResource = fatherResources
	}

	for _, st := range t.Tasks {
		ans = append(ans, st.GetAllTasksAndSubTasksList(fullName, ans[0].AllocatedResource)...)
	}

	return ans
}

func (t *Task) GetTaskByFullName(fullname string) (*Task, error) {
	var ans *Task = nil

	leafs := strings.Split(fullname, ".")
	if len(leafs) == 1 {
		if leafs[0] == t.Name {
			return t, nil
		} else {
			return nil, errors.New("Invalid task name " + fullname)
		}
	}

	mainActivity := leafs[1]

	for idx, st := range t.Tasks {
		if st.Name == mainActivity {
			ans = &t.Tasks[idx]
			break
		}
	}

	if ans == nil {
		return nil, errors.New("No task found with leaf " + mainActivity)
	}

	return ans.GetTaskByFullName(strings.Join(leafs[1:], "."))
}

func (t *Task) GetSubTasks() *[]Task {
	return &t.Tasks
}

func (t *Task) HasFlag(flag string) bool {
	for _, f := range t.Flags {
		if f == flag {
			return true
		}
	}

	return false
}

func (t *Task) HasLabelKey(key string) bool {
	for k, _ := range t.Labels {
		if k == key {
			return true
		}
	}
	return false
}

func (t *Task) Validate(ignoreError bool) error {
	if strings.Contains(t.Name, ".") {
		if !ignoreError {
			return errors.New("Invalid task name " + t.Name)
		}
		fmt.Println("Warning: Invalid task name " + t.Name)
	}

	if len(t.Tasks) > 0 {
		for _, st := range t.Tasks {
			err := st.Validate(ignoreError)
			if err != nil {
				if !ignoreError {
					return err
				}
			}
		}

	}

	return nil
}
