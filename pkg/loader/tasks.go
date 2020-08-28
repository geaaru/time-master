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
	specs "github.com/geaaru/time-master/pkg/specs"
)

func (i *TimeMasterInstance) GetTasks(opts specs.TaskResearch) ([]specs.Task, error) {
	ans := []specs.Task{}

	for _, client := range *i.GetClients() {

		if len(opts.Clients) > 0 && !regexEntry(client.GetName(), opts.Clients) {
			continue
		}

		for _, activity := range *client.GetActivities() {

			if opts.OnlyClosedActivity {
				if !activity.IsClosed() {
					continue
				}
			} else if !opts.ClosedActivity && activity.IsClosed() {
				continue
			}

			if len(opts.ActivityLabels) > 0 {
				matchLabel := false
				for k, v := range activity.Labels {
					if regexEntry(k+"="+v, opts.ActivityLabels) {
						matchLabel = true
						break
					}
				}
				if !matchLabel {
					continue
				}
			}

			if len(opts.ActivityFlags) > 0 {
				matchFlags := false
				for _, flag := range activity.Flags {
					if regexEntry(flag, opts.ActivityFlags) {
						matchFlags = true
						break
					}
				}
				if !matchFlags {
					continue
				}
			}

			for _, task := range activity.GetAllTasksList() {

				if opts.OnlyMilestone && task.Milestone == "" {
					continue
				} else if !opts.OnlyMilestone && !opts.Milestone && task.Milestone != "" {
					continue
				}

				if opts.WithEffort && task.Effort == "" {
					continue
				}

				if len(opts.Users) > 0 {
					matchUser := false
					for _, user := range task.AllocatedResource {
						if matchEntry(user, opts.Users) {
							matchUser = true
							break
						}
					}

					if !matchUser {
						continue
					}

				}

				if len(opts.Labels) > 0 {
					matchLabel := false
					for k, v := range task.Labels {
						if regexEntry(k+"="+v, opts.Labels) {
							matchLabel = true
							break
						}
					}
					if !matchLabel {
						continue
					}
				}

				if len(opts.Flags) > 0 {
					matchFlags := false
					for _, flag := range task.Flags {
						if regexEntry(flag, opts.Flags) {
							matchFlags = true
							break
						}
					}
					if !matchFlags {
						continue
					}
				}

				if len(opts.Tasks) > 0 && !regexEntry(task.Name, opts.Tasks) {
					continue
				}

				ans = append(ans, task)

			}

		}
	}

	return ans, nil
}
