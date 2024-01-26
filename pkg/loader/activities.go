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
	tmtime "github.com/geaaru/time-master/pkg/time"
	tools "github.com/geaaru/time-master/pkg/tools"
)

func (i *TimeMasterInstance) CalculateActivityBusinessProgress(aname string) (float64, error) {
	ans := float64(0)

	activity, _, err := i.GetActivityByName(aname)
	if err != nil {
		return ans, err
	}

	tasks := activity.GetAllTasksList()

	// Creating the rta map with the work time
	rOpts := specs.TimesheetResearch{
		ByTask:     true,
		IgnoreTime: true,
	}
	rtaMap, err := i.GetAggregatedTimesheetsMap(
		rOpts, "", "", []string{}, []string{aname},
	)
	if err != nil {
		return ans, err
	}

	totEffort := int64(0)
	totEffectiveWorked := int64(0)

	for idx := range tasks {

		tSecs := int64(0)

		if tasks[idx].Effort != "" {
			tSecs, err = tmtime.ParseDuration(
				tasks[idx].Effort, i.Config.GetWork().WorkHours,
			)
			if err != nil {
				return ans, err
			}
		}

		rta, ok := rtaMap[tasks[idx].Name]
		if ok {

			// If the worked effort is greather than effort i using all worked
			// effort.
			if tasks[idx].IsCompleted() || rta.Seconds > tSecs {
				totEffort += rta.Seconds
				totEffectiveWorked += rta.Seconds
			} else {
				totEffort += tSecs
				totEffectiveWorked += rta.Seconds
			}
		} else if tSecs > 0 {

			// POST: no hours worked
			totEffort += tSecs
			if tasks[idx].IsCompleted() {
				totEffectiveWorked += tSecs
			}
		}
	}

	if totEffort > 0 && totEffectiveWorked > 0 {
		ans = (float64(totEffectiveWorked) * 100) / float64(totEffort)
	}

	return ans, err
}

func (i *TimeMasterInstance) GetActivities(opts specs.ActivityResearch) ([]specs.Activity, error) {
	ans := []specs.Activity{}

	for _, client := range *i.GetClients() {

		if len(opts.Clients) > 0 && !tools.MatchEntry(client.GetName(), opts.Clients) {
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

			if len(opts.ExcludeNames) > 0 {
				if tools.RegexEntry(activity.Name, opts.ExcludeNames) {
					continue
				}
			}

			if len(opts.ExcludeFlags) > 0 {
				matchFlags := false
				for _, flag := range activity.Flags {
					if tools.RegexEntry(flag, opts.ExcludeFlags) {
						matchFlags = true
						break
					}
				}

				if matchFlags {
					continue
				}

			}

			if len(opts.Labels) > 0 {
				if len(activity.Labels) == 0 {
					continue
				}

				if opts.LabelsInAnd {
					skipActivity := false
					for _, label := range opts.Labels {

						matchLabel := false
						l := []string{label}

						for k, v := range activity.Labels {
							if tools.RegexEntry(k+"="+v, l) {
								matchLabel = true
								break
							}
						}

						if !matchLabel {
							skipActivity = true
							break
						}
					}
					if skipActivity {
						continue
					}
				} else {

					matchLabel := false
					for _, flag := range activity.Labels {
						if tools.RegexEntry(flag, opts.Labels) {
							matchLabel = true
							break
						}
					}
					if !matchLabel {
						continue
					}
				}
			}

			if len(opts.Flags) > 0 {
				if len(activity.Flags) == 0 {
					continue
				}

				matchFlags := false
				for _, flag := range activity.Flags {
					if tools.RegexEntry(flag, opts.Flags) {
						matchFlags = true
						break
					}
				}
				if !matchFlags {
					continue
				}
			}

			if len(opts.Names) > 0 {
				if !tools.RegexEntry(activity.Name, opts.Names) {
					continue
				}
			}

			ans = append(ans, activity)

		}

	}

	return ans, nil
}
