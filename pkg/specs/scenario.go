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

	"gopkg.in/yaml.v2"

	time "github.com/geaaru/time-master/pkg/time"
)

func ScenarioFromYaml(data []byte, file string) (*Scenario, error) {
	ans := &Scenario{}
	if err := yaml.Unmarshal(data, ans); err != nil {
		return nil, err
	}
	ans.File = file

	return ans, nil
}

func (s *Scenario) SetNow(n string) {
	s.NowTime = n
}

func (s *Scenario) GetResourceCost4Date(dstr, resourceUser string) (float64, error) {
	ans := float64(0)
	notFound := true

	d, err := time.ParseTimestamp(dstr, true)
	if err != nil {
		return ans, err
	}

	for _, r := range s.ResourceCosts {
		if r.User != resourceUser {
			continue
		}

		startTime, err := time.ParseTimestamp(r.Period.StartPeriod, true)
		if err != nil {
			return ans, err
		}

		if startTime.Unix() > d.Unix() {
			continue
		}

		if r.Period.EndPeriod != "" {
			endTime, err := time.ParseTimestamp(r.Period.EndPeriod, true)
			if err != nil {
				return ans, err
			}

			if endTime.Unix() <= d.Unix() {
				continue
			}
		}

		notFound = false
		ans = r.Cost
		break
	}

	if notFound {
		return ans, errors.New(
			fmt.Sprintf("No resource cost found for user %s and period %s",
				resourceUser, dstr))
	}

	return ans, nil
}

func (s *Scenario) GetResourceRate4Date(dstr, resourceUser string) (float64, error) {

	ans := float64(0)
	notFound := true

	d, err := time.ParseTimestamp(dstr, true)
	if err != nil {
		return ans, err
	}

	for _, r := range s.Rates {
		if r.User != resourceUser {
			continue
		}

		startTime, err := time.ParseTimestamp(r.Period.StartPeriod, true)
		if err != nil {
			return ans, err
		}

		if startTime.Unix() > d.Unix() {
			continue
		}

		if r.Period.EndPeriod != "" {
			endTime, err := time.ParseTimestamp(r.Period.EndPeriod, true)
			if err != nil {
				return ans, err
			}

			if endTime.Unix() <= d.Unix() {
				continue
			}
		}

		notFound = false
		ans = r.Rate
		break
	}

	if notFound {
		return ans, errors.New(
			fmt.Sprintf("No resource rate found for user %s and period %s",
				resourceUser, dstr))
	}

	return ans, nil
}
