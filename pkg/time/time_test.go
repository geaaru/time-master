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
package time_test

import (
	. "github.com/geaaru/time-master/pkg/time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Time Test", func() {

	Context("Day duration", func() {

		It("Convert 1d", func() {
			sec, err := ParseDuration("1d", 8)

			Expect(err).Should(BeNil())
			Expect(sec).To(Equal(int64(60 * 60 * 8)))
		})

		It("Convert 0.5d", func() {
			sec, err := ParseDuration("0.5d", 8)

			Expect(err).Should(BeNil())
			Expect(sec).To(Equal(int64(60 * 60 * 4)))
		})

		It("Convert 15m", func() {
			sec, err := ParseDuration("15m", 8)

			Expect(err).Should(BeNil())
			Expect(sec).To(Equal(int64(60 * 15)))
		})

		It("Convert 15h", func() {
			sec, err := ParseDuration("15h", 8)

			Expect(err).Should(BeNil())
			Expect(sec).To(Equal(int64(60 * 60 * 15)))
		})

		It("Convert 1h", func() {
			sec, err := ParseDuration("1h", 8)

			Expect(err).Should(BeNil())
			Expect(sec).To(Equal(int64(60 * 60 * 1)))
		})
	})

	Context("Seconds converter", func() {

		It("Convert 1h", func() {
			duration, err := Seconds2Duration(int64(60 * 60))

			Expect(err).Should(BeNil())
			Expect(duration).To(Equal("1h"))
		})

		It("Convert 1h15m", func() {
			duration, err := Seconds2Duration(int64(60*60 + 60*15))

			Expect(err).Should(BeNil())
			Expect(duration).To(Equal("1h15m"))
		})

		It("Convert 10h", func() {
			duration, err := Seconds2Duration(int64(60 * 60 * 10))

			Expect(err).Should(BeNil())
			Expect(duration).To(Equal("10h"))
		})

		It("Convert 10h35m10s", func() {
			duration, err := Seconds2Duration(int64(60*60*10 + 60*35 + 10))

			Expect(err).Should(BeNil())
			Expect(duration).To(Equal("10h35m10s"))
		})

		It("Handle invalid sec", func() {
			_, err := Seconds2Duration(int64(-1))

			Expect(err).ShouldNot(BeNil())
		})
	})
})
