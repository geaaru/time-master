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
	ProfitPerc string `json:"profic_perc,omitempty" yaml:"profic_perc,omitempty"`

	Effort   int64 `json:"effort_sec,omitempty" yaml:"effort_sec,omitempty"`
	WorkSecs int64 `json:"work_sec,omitempty" yaml:"work_sec,omitempty"`
}
