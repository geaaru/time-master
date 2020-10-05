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
	. "github.com/geaaru/time-master/pkg/specs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resource Test", func() {

	Context("Available", func() {

		It("Test 1", func() {

			resource := NewResource("geaaru", "User One")
			resource.AddHoliday(
				ResourceHolidays{
					Period: &Period{
						StartPeriod: "2020-01-01",
						EndPeriod:   "2020-01-03",
					},
				},
			)

			resource.AddSick(
				ResourceSick{
					Period: &Period{
						StartPeriod: "2020-02-01",
						EndPeriod:   "2020-02-03",
					},
				},
			)

			resource.AddUnemployed(
				ResourceUnemployed{
					Period: &Period{
						StartPeriod: "2020-03-01",
						EndPeriod:   "2020-03-04",
					},
				},
			)

			available, err := resource.IsAvailable("2020-01-10")
			Expect(err).Should(BeNil())
			Expect(available).Should(Equal(true))
			available, err = resource.IsAvailable("2020-02-02")
			Expect(err).Should(BeNil())
			Expect(available).Should(Equal(false))
			available, err = resource.IsAvailable("2020-03-01")
			Expect(err).Should(BeNil())
			Expect(available).Should(Equal(false))
			available, err = resource.IsAvailable("2020-03-04")
			Expect(err).Should(BeNil())
			Expect(available).Should(Equal(false))

		})

		It("Test 2", func() {

			resource := NewResource("geaaru", "User One")
			resource.AddHoliday(
				ResourceHolidays{
					Period: &Period{
						StartPeriod: "2020-01-01",
						EndPeriod:   "2020-01-03",
					},
				},
			)

			resource.AddSick(
				ResourceSick{
					Period: &Period{
						StartPeriod: "2020-02-01",
						EndPeriod:   "2020-02-03",
					},
				},
			)

			resource.AddUnemployed(
				ResourceUnemployed{
					Period: &Period{
						StartPeriod: "2020-03-01",
					},
				},
			)

			available, err := resource.IsAvailable("2020-03-10")
			Expect(err).Should(BeNil())
			Expect(available).Should(Equal(false))
			available, err = resource.IsAvailable("2019-01-01")
			Expect(err).Should(BeNil())
			Expect(available).Should(Equal(true))
			available, err = resource.IsAvailable("2020-06-10")
			Expect(err).Should(BeNil())
			Expect(available).Should(Equal(false))
		})

	})

})
