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

func initConfig() *specs.TimeMasterConfig {
	config := specs.NewTimeMasterConfig(nil)
	config.GetGeneral().Debug = true
	config.GetWork().WorkHours = 8
	config.GetWork().TaskDefaultPriority = 100
	return config
}

func initializeScheduler(config *specs.TimeMasterConfig) *SimpleScheduler {
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

func initializeSchedulerMultiUser(config *specs.TimeMasterConfig) *SimpleScheduler {
	scenario := &specs.Scenario{
		Name:      "test",
		Scheduler: "simple",
	}

	client := specs.NewClient("TEST1")
	activity := specs.NewActivity("ACTIVITY1", "")

	task := specs.NewTask("TASK1", "", "", []string{"user1", "user2", "user3"})
	task.Recursive.Enable = true
	task.Recursive.Mode = "weekly"
	task.Recursive.Duration = "12h"

	activity.AddTask(task)
	client.AddActivity(*activity)

	scheduler := NewSimpleScheduler(config, scenario)
	scheduler.Resources = []specs.Resource{
		*specs.NewResource("user1", "User One"),
		*specs.NewResource("user2", "User Two"),
		*specs.NewResource("user3", "User Two"),
	}
	scheduler.Timesheets = []specs.AgendaTimesheets{}
	scheduler.Clients = []specs.Client{*client}

	return scheduler
}

var _ = Describe("Scheduler Recursive Test", func() {

	config := initConfig()

	Context("Daily recursive task", func() {

		scheduler := initializeScheduler(config)
		scheduler.Clients[0].Activities[0].Tasks[0].Period.EndPeriod = "2020-09-12"

		scheduler.CreateTaskScheduled()
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

		scheduler := initializeScheduler(config)
		scheduler.Clients[0].Activities[0].Tasks[0].Period.EndPeriod = "2020-11-01"
		scheduler.Clients[0].Activities[0].Tasks[0].Recursive.Mode = "monthly"

		scheduler.CreateTaskScheduled()
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

		scheduler := initializeScheduler(config)
		scheduler.Clients[0].Activities[0].Tasks[0].Period.EndPeriod = "2020-10-01"
		scheduler.Clients[0].Activities[0].Tasks[0].Recursive.Mode = "weekly"

		scheduler.CreateTaskScheduled()
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

	Context("Weekly recursive task multiple users", func() {

		scheduler := initializeSchedulerMultiUser(config)
		scheduler.Clients[0].Activities[0].Tasks[0].Period.EndPeriod = "2020-10-01"
		scheduler.Clients[0].Activities[0].Tasks[0].Recursive.Mode = "weekly"

		scheduler.CreateTaskScheduled()
		scheduler.Init()
		seer := NewRecursiveTaskSeer(scheduler, &scheduler.Scenario.Schedule[0])
		err := seer.DoPrevision("2020-09-06")

		prevision := scheduler.Scenario

		It("Weekly 12h", func() {
			Expect(err).Should(BeNil())
			Expect(len(prevision.Schedule[0].Timesheets)).To(Equal(8))
			Expect(prevision.Schedule[0].Timesheets).To(Equal(
				[]specs.ResourceTimesheet{

					*specs.NewResourceTimesheet("user1", "2020-09-07", "ACTIVITY1.TASK1", "28800s"),
					*specs.NewResourceTimesheet("user2", "2020-09-07", "ACTIVITY1.TASK1", "14400s"),
					*specs.NewResourceTimesheet("user1", "2020-09-14", "ACTIVITY1.TASK1", "28800s"),
					*specs.NewResourceTimesheet("user2", "2020-09-14", "ACTIVITY1.TASK1", "14400s"),

					*specs.NewResourceTimesheet("user1", "2020-09-21", "ACTIVITY1.TASK1", "28800s"),
					*specs.NewResourceTimesheet("user2", "2020-09-21", "ACTIVITY1.TASK1", "14400s"),
					*specs.NewResourceTimesheet("user1", "2020-09-28", "ACTIVITY1.TASK1", "28800s"),
					*specs.NewResourceTimesheet("user2", "2020-09-28", "ACTIVITY1.TASK1", "14400s"),
				},
			))
		})

	})

	Context("Monthly recursive task multiple users", func() {

		scheduler := initializeSchedulerMultiUser(config)
		scheduler.Clients[0].Activities[0].Tasks[0].Period.EndPeriod = "2020-11-15"
		scheduler.Clients[0].Activities[0].Tasks[0].Recursive.Mode = "monthly"
		scheduler.Clients[0].Activities[0].Tasks[0].Recursive.Duration = "7d"

		scheduler.CreateTaskScheduled()
		scheduler.Init()
		seer := NewRecursiveTaskSeer(scheduler, &scheduler.Scenario.Schedule[0])
		err := seer.DoPrevision("2020-09-06")

		prevision := scheduler.Scenario

		It("Monthly 7d", func() {
			Expect(err).Should(BeNil())
			Expect(len(prevision.Schedule[0].Timesheets)).To(Equal(21))
			Expect(prevision.Schedule[0].Timesheets).To(Equal(
				[]specs.ResourceTimesheet{

					*specs.NewResourceTimesheet("user1", "2020-09-07", "ACTIVITY1.TASK1", "28800s"),
					*specs.NewResourceTimesheet("user2", "2020-09-07", "ACTIVITY1.TASK1", "28800s"),
					*specs.NewResourceTimesheet("user3", "2020-09-07", "ACTIVITY1.TASK1", "28800s"),
					*specs.NewResourceTimesheet("user1", "2020-09-08", "ACTIVITY1.TASK1", "28800s"),
					*specs.NewResourceTimesheet("user2", "2020-09-08", "ACTIVITY1.TASK1", "28800s"),
					*specs.NewResourceTimesheet("user3", "2020-09-08", "ACTIVITY1.TASK1", "28800s"),
					*specs.NewResourceTimesheet("user1", "2020-09-09", "ACTIVITY1.TASK1", "28800s"),

					*specs.NewResourceTimesheet("user1", "2020-10-01", "ACTIVITY1.TASK1", "28800s"),
					*specs.NewResourceTimesheet("user2", "2020-10-01", "ACTIVITY1.TASK1", "28800s"),
					*specs.NewResourceTimesheet("user3", "2020-10-01", "ACTIVITY1.TASK1", "28800s"),
					*specs.NewResourceTimesheet("user1", "2020-10-02", "ACTIVITY1.TASK1", "28800s"),
					*specs.NewResourceTimesheet("user2", "2020-10-02", "ACTIVITY1.TASK1", "28800s"),
					*specs.NewResourceTimesheet("user3", "2020-10-02", "ACTIVITY1.TASK1", "28800s"),
					*specs.NewResourceTimesheet("user1", "2020-10-05", "ACTIVITY1.TASK1", "28800s"),

					*specs.NewResourceTimesheet("user1", "2020-11-02", "ACTIVITY1.TASK1", "28800s"),
					*specs.NewResourceTimesheet("user2", "2020-11-02", "ACTIVITY1.TASK1", "28800s"),
					*specs.NewResourceTimesheet("user3", "2020-11-02", "ACTIVITY1.TASK1", "28800s"),
					*specs.NewResourceTimesheet("user1", "2020-11-03", "ACTIVITY1.TASK1", "28800s"),
					*specs.NewResourceTimesheet("user2", "2020-11-03", "ACTIVITY1.TASK1", "28800s"),
					*specs.NewResourceTimesheet("user3", "2020-11-03", "ACTIVITY1.TASK1", "28800s"),
					*specs.NewResourceTimesheet("user1", "2020-11-04", "ACTIVITY1.TASK1", "28800s"),
				},
			))
		})

	})
})
