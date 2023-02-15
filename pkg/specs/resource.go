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

	time "github.com/geaaru/time-master/pkg/time"

	"gopkg.in/yaml.v3"
)

func ResourceFromYaml(data []byte, file string) (*Resource, error) {
	ans := &Resource{}
	if err := yaml.Unmarshal(data, ans); err != nil {
		return nil, err
	}
	ans.File = file

	return ans, nil
}

func NewResource(user, name string) *Resource {
	return &Resource{
		User:       user,
		Name:       name,
		Email:      []string{},
		Phone:      []string{},
		Holidays:   []ResourceHolidays{},
		Sick:       []ResourceSick{},
		Unemployed: []ResourceUnemployed{},
	}
}

func (r *Resource) AddHoliday(rh ResourceHolidays) {
	r.Holidays = append(r.Holidays, rh)
}

func (r *Resource) AddSick(s ResourceSick) { r.Sick = append(r.Sick, s) }

func (r *Resource) AddUnemployed(ru ResourceUnemployed) {
	r.Unemployed = append(r.Unemployed, ru)
}

func (r *Resource) IsAvailable(workDate string) (bool, error) {
	wTime, err := time.ParseTimestamp(workDate, true)
	if err != nil {
		return false, err
	}

	if len(r.Holidays) > 0 {
		for _, h := range r.Holidays {

			startTime, err := time.ParseTimestamp(h.Period.StartPeriod, true)
			if err != nil {
				return false, err
			}

			if wTime.Unix() < startTime.Unix() {
				// POST: the date is before the holiday
				continue
			}

			if wTime.Unix() == startTime.Unix() || h.Period.EndPeriod == "" {
				return false, nil
			}

			endTime, err := time.ParseTimestamp(h.Period.EndPeriod, true)
			if err != nil {
				return false, err
			}

			if wTime.Unix() <= endTime.Unix() {
				return false, nil
			}
		}
	}

	if len(r.Sick) > 0 {
		for _, e := range r.Sick {
			startTime, err := time.ParseTimestamp(e.Period.StartPeriod, true)
			if err != nil {
				return false, err
			}
			if wTime.Unix() < startTime.Unix() {
				// POST: the date is before the holiday
				continue
			}

			if wTime.Unix() == startTime.Unix() || e.Period.EndPeriod == "" {
				return false, nil
			}

			endTime, err := time.ParseTimestamp(e.Period.EndPeriod, true)
			if err != nil {
				return false, err
			}

			if wTime.Unix() <= endTime.Unix() {
				return false, nil
			}
		}
	}

	if len(r.Unemployed) > 0 {
		for _, e := range r.Unemployed {
			startTime, err := time.ParseTimestamp(e.Period.StartPeriod, true)
			if err != nil {
				return false, err
			}

			if wTime.Unix() < startTime.Unix() {
				// POST: the date is before the holiday
				continue
			}

			if wTime.Unix() == startTime.Unix() || e.Period.EndPeriod == "" {
				return false, nil
			}

			endTime, err := time.ParseTimestamp(e.Period.EndPeriod, true)
			if err != nil {
				return false, err
			}

			if wTime.Unix() <= endTime.Unix() {
				return false, nil
			}
		}
	}

	return true, nil
}

func (r *Resource) Validate() error {
	if len(r.Holidays) > 0 {
		for _, h := range r.Holidays {
			err := r.validatePeriod(h.Period, "holiday period")
			if err != nil {
				return err
			}
		}
	}

	if len(r.Sick) > 0 {
		for _, s := range r.Sick {
			err := r.validatePeriod(s.Period, "sick")
			if err != nil {
				return err
			}
		}
	}

	if len(r.Unemployed) > 0 {
		withOpenUnemployed := false
		for _, u := range r.Unemployed {
			err := r.validatePeriod(u.Period, "unemployed")
			if err != nil {
				return err
			}
			if u.Period.EndPeriod == "" {
				if withOpenUnemployed {
					return errors.New("Multiple unemployed entry without end period")
				}
				withOpenUnemployed = true
			}
		}
	}

	return nil
}

func (r *Resource) validatePeriod(p *Period, note string) error {
	if p == nil {
		return errors.New("Invalid " + note + " entry")
	}
	if p.StartPeriod == "" || p.EndPeriod == "" {
		return errors.New("Invalid " + note + " entry")
	}

	if p.StartPeriod != "" {
		_, err := time.ParseTimestamp(p.StartPeriod, true)
		if err != nil {
			return errors.New(
				fmt.Sprintf("Invalid date %s (%s) on resource %s: %s",
					p.StartPeriod, note, r.User, err.Error()))
		}
	}

	if p.EndPeriod != "" {
		_, err := time.ParseTimestamp(p.StartPeriod, true)
		if err != nil {
			return errors.New(
				fmt.Sprintf("Invalid date %s (%s) on resource %s: %s",
					p.StartPeriod, note, r.User, err.Error()))
		}
	}

	return nil
}
