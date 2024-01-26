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
package specs_test

import (
	"time"

	. "github.com/geaaru/time-master/pkg/specs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Agenda Timesheets Test", func() {

	Context("Last Date Seconds", func() {

		It("Last Date 1", func() {

			agenda := AgendaTimesheets{
				Name: "test",
				Timesheets: []ResourceTimesheet{
					{
						Period: &Period{
							StartPeriod: "2020-01-01",
						},
						User:     "user1",
						Task:     "ACTIVITY1.task1",
						Duration: "4d",
					},
					{
						Period: &Period{
							StartPeriod: "2020-01-04",
						},
						User:     "user1",
						Task:     "ACTIVITY1.task1",
						Duration: "4d",
					},
				},
			}

			expectedDate, err := time.Parse("2006-01-02", "2020-01-04")

			Expect(err).Should(BeNil())
			aLastDate, err := agenda.GetLastDateSecs(true)
			Expect(err).Should(BeNil())
			Expect(aLastDate).To(Equal(expectedDate.Unix()))
		})

	})

})
