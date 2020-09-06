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
package scheduler_test

import (
	. "github.com/geaaru/time-master/pkg/scheduler"
	specs "github.com/geaaru/time-master/pkg/specs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func initializeScheduler() *SimpleScheduler {
	config := specs.NewTimeMasterConfig(nil)
	config.GetGeneral().Debug = true
	config.GetWork().WorkHours = 8
	config.GetWork().TaskDefaultPriority = 100
	scenario := &specs.Scenario{
		Name:      "test",
		Scheduler: "simple",
	}

	client := specs.NewClient("TEST1")
	activity := specs.NewActivity("ACTIVITY1", "")

	task := specs.NewTask("TASK1", "", "", []string{"user1"})
	task.Recursive.Enable = true
	task.Recursive.Mode = "daily"
	task.Recursive.Duration = "2h"

	activity.AddTask(task)
	client.AddActivity(*activity)

	scheduler := NewSimpleScheduler(config, scenario)
	scheduler.Resources = []specs.Resource{
		*specs.NewResource("user1", "User One"),
	}
	scheduler.Timesheets = []specs.AgendaTimesheets{}
	scheduler.Clients = []specs.Client{*client}

	return scheduler
}

var _ = Describe("Scheduler Recursive Test", func() {

	Context("Daily recursive task", func() {

		scheduler := initializeScheduler()
		scheduler.Clients[0].Activities[0].Tasks[0].Period.EndPeriod = "2020-09-12"

		scheduler.Init()

		seer := NewRecursiveTaskSeer(scheduler, &scheduler.Scenario.Schedule[0])
		err := seer.DoPrevision("2020-09-06")

		prevision := scheduler.Scenario

		It("Test daily 2h", func() {
			Expect(err).Should(BeNil())
			Expect(len(prevision.Schedule[0].Timesheets)).To(Equal(5))

			Expect(prevision.Schedule[0].Timesheets).To(Equal(
				[]specs.ResourceTimesheet{

					*specs.NewResourceTimesheet("user1", "2020-09-07", "ACTIVITY1.TASK1", "7200s"),
					*specs.NewResourceTimesheet("user1", "2020-09-08", "ACTIVITY1.TASK1", "7200s"),
					*specs.NewResourceTimesheet("user1", "2020-09-09", "ACTIVITY1.TASK1", "7200s"),
					*specs.NewResourceTimesheet("user1", "2020-09-10", "ACTIVITY1.TASK1", "7200s"),
					*specs.NewResourceTimesheet("user1", "2020-09-11", "ACTIVITY1.TASK1", "7200s"),
				},
			))
		})

	})

	Context("Daily recursive task", func() {

		scheduler := initializeScheduler()
		scheduler.Clients[0].Activities[0].Tasks[0].Period.EndPeriod = "2020-11-01"
		scheduler.Clients[0].Activities[0].Tasks[0].Recursive.Mode = "monthly"

		scheduler.Init()
		seer := NewRecursiveTaskSeer(scheduler, &scheduler.Scenario.Schedule[0])
		err := seer.DoPrevision("2020-09-06")

		prevision := scheduler.Scenario

		It("Monthly 2h", func() {
			Expect(err).Should(BeNil())
			Expect(len(prevision.Schedule[0].Timesheets)).To(Equal(2))
			Expect(prevision.Schedule[0].Timesheets).To(Equal(
				[]specs.ResourceTimesheet{

					*specs.NewResourceTimesheet("user1", "2020-09-07", "ACTIVITY1.TASK1", "7200s"),
					*specs.NewResourceTimesheet("user1", "2020-10-01", "ACTIVITY1.TASK1", "7200s"),
				},
			))
		})

	})

	Context("Weekly recursive task", func() {

		scheduler := initializeScheduler()
		scheduler.Clients[0].Activities[0].Tasks[0].Period.EndPeriod = "2020-10-01"
		scheduler.Clients[0].Activities[0].Tasks[0].Recursive.Mode = "weekly"

		scheduler.Init()
		seer := NewRecursiveTaskSeer(scheduler, &scheduler.Scenario.Schedule[0])
		err := seer.DoPrevision("2020-09-06")

		prevision := scheduler.Scenario

		It("Monthly 2h", func() {
			Expect(err).Should(BeNil())
			Expect(len(prevision.Schedule[0].Timesheets)).To(Equal(4))
			Expect(prevision.Schedule[0].Timesheets).To(Equal(
				[]specs.ResourceTimesheet{

					*specs.NewResourceTimesheet("user1", "2020-09-07", "ACTIVITY1.TASK1", "7200s"),
					*specs.NewResourceTimesheet("user1", "2020-09-14", "ACTIVITY1.TASK1", "7200s"),
					*specs.NewResourceTimesheet("user1", "2020-09-21", "ACTIVITY1.TASK1", "7200s"),
					*specs.NewResourceTimesheet("user1", "2020-09-28", "ACTIVITY1.TASK1", "7200s"),
				},
			))
		})

	})
})
