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
	"fmt"

	time "github.com/geaaru/time-master/pkg/time"
)

func NewActivityReport(a Activity, filtered bool) *ActivityReport {
	ans := &ActivityReport{
		Activity: &Activity{
			Name:         a.Name,
			Description:  a.Description,
			Note:         a.Note,
			Priority:     a.Priority,
			File:         a.File,
			Disabled:     a.Disabled,
			Closed:       a.Closed,
			Offer:        a.Offer,
			TimeMaterial: a.TimeMaterial,
			Labels:       a.Labels,
			Flags:        a.Flags,
			Tasks:        []Task{},
		},
	}

	if !filtered {
		ans.TMDailyOffer = a.TMDailyOffer
		for _, task := range a.Tasks {
			ans.Tasks = append(ans.Tasks, *task.Clone(filtered))
		}
	}

	return ans
}

func (a *ActivityReport) SetRevenuePlan(v float64) { a.RevenuePlan = v }
func (a *ActivityReport) SetCost(v float64)        { a.Cost = v }
func (a *ActivityReport) SetProfit(v float64)      { a.Profit = v }
func (a *ActivityReport) SetWork(v string)         { a.Work = v }
func (a *ActivityReport) SetWorkPerc(v string)     { a.WorkPerc = v }
func (a *ActivityReport) SetEffort(v int64)        { a.Effort = v }
func (a *ActivityReport) SetWorkSecs(v int64)      { a.WorkSecs = v }

func (a *ActivityReport) CalculateWorkPerc() {
	if a.WorkSecs > 0 && a.Effort > 0 {
		a.WorkPerc = fmt.Sprintf("%02.02f", (float64(a.WorkSecs)/float64(a.Effort))*100)
	}
}

func (a *ActivityReport) CalculateProfitPerc() {
	if a.Offer > 0 && a.WorkSecs > 0 {
		a.ProfitPerc = fmt.Sprintf("%02.02f", ((float64(a.Profit) * 100) / float64(a.Offer)))
	} else if a.IsTimeAndMaterial() && a.Profit > 0 {
		a.ProfitPerc = fmt.Sprintf("%02.02f", ((float64(a.Profit) * 100) / float64(a.RevenuePlan)))
	}
}

func (a *ActivityReport) CalculateDuration() error {
	d, err := time.Seconds2Duration(a.Effort)
	if err != nil {
		return err
	}

	a.Duration = d
	return nil
}

func (a *ActivityReport) GetDuration() string {
	if a.Duration == "" && a.Effort > 0 {
		a.CalculateDuration()
	}
	return a.Duration
}
