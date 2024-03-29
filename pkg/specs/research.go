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

type TimesheetResearch struct {
	ByUser     bool
	ByTask     bool
	ByActivity bool
	Monthly    bool
	IgnoreTime bool
}

type TaskResearch struct {
	ClosedActivity     bool
	OnlyClosedActivity bool
	Milestone          bool
	OnlyMilestone      bool
	WithEffort         bool
	Users              []string
	Flags              []string
	ActivityFlags      []string
	Labels             []string
	ActivityLabels     []string
	Tasks              []string
	Clients            []string
	ExcludeFlags       []string
}

type ActivityResearch struct {
	ClosedActivity     bool
	OnlyClosedActivity bool
	Flags              []string
	Labels             []string
	Clients            []string
	Names              []string
	ExcludeNames       []string
	ExcludeFlags       []string
	LabelsInAnd        bool
}
