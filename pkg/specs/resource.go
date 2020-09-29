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

	"gopkg.in/yaml.v2"
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

func (r *Resource) Validate() error {
	if len(r.Holidays) > 0 {
		for _, h := range r.Holidays {
			if h.Period == nil {
				return errors.New("Invalid holiday period entry")
			}
			if h.Period.StartPeriod == "" || h.Period.EndPeriod == "" {
				return errors.New("Invalid holiday entry")
			}
		}
	}

	if len(r.Sick) > 0 {
		for _, s := range r.Sick {
			if s.Period == nil {
				return errors.New("Invalid sick period entry")
			}
			if s.Period.StartPeriod == "" || s.Period.EndPeriod == "" {
				return errors.New("Invalid sick entry")
			}
		}
	}

	if len(r.Unemployed) > 0 {
		withOpenUnemployed := false
		for _, u := range r.Unemployed {
			if u.Period == nil {
				return errors.New("Invalid unemployed period entry")
			}

			if u.Period.StartPeriod == "" {
				return errors.New("Invalid unemployed entry with empty start period")
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
