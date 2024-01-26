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
package gantt

import (
	"encoding/json"
	"sort"
	"strings"

	specs "github.com/geaaru/time-master/pkg/specs"
	time "github.com/geaaru/time-master/pkg/time"
)

// Json format of the task required by the project: https://github.com/frappe/gantt
// Thanks for this project.
type FrappeGanttTask struct {
	Id           string  `json:"id"`
	Name         string  `json:"name"`
	Start        string  `json:"start"`
	End          string  `json:"end"`
	Progress     float64 `json:"progress,omitempty"`
	Dependencies string  `json:"dependencies,omitempty"`
	CustomClass  string  `json:"custom_class,omitempty"`
	StartTime    int64   `json:"-"`
	EndTime      int64   `json:"-"`
}

type FGTaskSorter []FrappeGanttTask
type FGTaskEndDateSorter []FrappeGanttTask

type FrappeGanttProducer struct {
	*DefaultGanttProducer
}

func NewFrappeGanttProducer(config *specs.TimeMasterConfig) *FrappeGanttProducer {
	return &FrappeGanttProducer{
		DefaultGanttProducer: newDefaultGanttProducer(config),
	}
}

func (f *FrappeGanttProducer) Build(s *specs.ScenarioSchedule, opts ProducerOpts) ([]byte, error) {
	tasks := []FrappeGanttTask{}
	ans := []byte{}

	for _, ts := range s.Schedule {

		startTime, err := time.ParseTimestamp(ts.Period.StartPeriod, true)
		if err != nil {
			f.Logger.Error("Error on on parse start date of task ", ts.Name)
			return ans, err
		}

		endTime, err := time.ParseTimestamp(ts.Period.EndPeriod, true)
		if err != nil {
			f.Logger.Error("Error on on parse end date of task ", ts.Name)
			return ans, err
		}
		ft := FrappeGanttTask{
			Id:        ts.Name,
			Start:     ts.Period.StartPeriod,
			End:       ts.Period.EndPeriod,
			StartTime: startTime.Unix(),
			EndTime:   endTime.Unix(),
		}

		words := strings.Split(ts.Name, ".")

		if opts.ShowActivityOnTasks {
			ft.Name = words[0] + " - " + ts.Description
		} else {
			ft.Name = ts.Description
		}

		if ts.Task.Milestone != "" {
			ft.CustomClass = "bar-milestone"
			ft.Name = words[0] + " - " + ts.Description
		} else {
			ft.Progress = ts.Progress
		}

		for idx, dep := range ts.Task.Depends {
			if idx == 0 {
				ft.Dependencies = dep
			} else {
				ft.Dependencies = ft.Dependencies + ", " + dep
			}
		}

		tasks = append(tasks, ft)
	}

	if opts.OrderByEndTime {
		sort.Sort(FGTaskEndDateSorter(tasks))
	} else {
		sort.Sort(FGTaskSorter(tasks))
	}

	return json.Marshal(tasks)
}

func (t FGTaskSorter) Len() int      { return len(t) }
func (t FGTaskSorter) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t FGTaskSorter) Less(i, j int) bool {
	if t[i].StartTime == t[j].StartTime {
		return t[i].CustomClass != ""
	}
	return t[i].StartTime < t[j].StartTime
}

func (t FGTaskEndDateSorter) Len() int      { return len(t) }
func (t FGTaskEndDateSorter) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t FGTaskEndDateSorter) Less(i, j int) bool {
	if t[i].EndTime == t[j].EndTime {
		return t[i].CustomClass != ""
	}
	return t[i].EndTime < t[j].EndTime
}
