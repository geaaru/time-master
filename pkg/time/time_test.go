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
	"time"

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

		It("Convert 3600s", func() {
			sec, err := ParseDuration("86400s", 8)

			Expect(err).Should(BeNil())
			Expect(sec).To(Equal(int64(60 * 60 * 24)))
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

		It("Parse timestamp1", func() {
			t, err := ParseTimestamp("2020-08-16", true)

			Expect(err).Should(BeNil())
			Expect(t.Year()).To(Equal(2020))
			Expect(t.Month()).To(Equal(time.August))
			Expect(t.Day()).To(Equal(16))
		})

		It("Parse timestamp2", func() {
			t, err := ParseTimestamp("2020-08-16 08:00:00", false)

			Expect(err).Should(BeNil())
			Expect(t.Year()).To(Equal(2020))
			Expect(t.Month()).To(Equal(time.August))
			Expect(t.Day()).To(Equal(16))
			Expect(t.Hour()).To(Equal(8))
			Expect(t.Minute()).To(Equal(0))
			Expect(t.Second()).To(Equal(0))
		})

		It("Parse timestamp3", func() {
			t, err := ParseTimestamp("2020-08-16 08:00:00", true)

			Expect(err).Should(BeNil())
			Expect(t.Year()).To(Equal(2020))
			Expect(t.Month()).To(Equal(time.August))
			Expect(t.Day()).To(Equal(16))
			Expect(t.Hour()).To(Equal(0))
			Expect(t.Minute()).To(Equal(0))
			Expect(t.Second()).To(Equal(0))
		})
	})

	Context("Next Work Day", func() {

		It("Parse1", func() {
			d1, err := GetNextWorkDay("2020-09-03")
			Expect(err).Should(BeNil())
			Expect(d1).To(Equal("2020-09-04"))
		})

		It("Parse2 - Friday", func() {
			d1, err := GetNextWorkDay("2020-09-04")
			Expect(err).Should(BeNil())
			Expect(d1).To(Equal("2020-09-07"))
		})

		It("Parse3 - Saturday", func() {
			d1, err := GetNextWorkDay("2020-09-05")
			Expect(err).Should(BeNil())
			Expect(d1).To(Equal("2020-09-07"))
		})

		It("Parse4 - Sunday", func() {
			d1, err := GetNextWorkDay("2020-09-06")
			Expect(err).Should(BeNil())
			Expect(d1).To(Equal("2020-09-07"))
		})
	})
})
