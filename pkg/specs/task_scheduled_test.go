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
	"sort"

	. "github.com/geaaru/time-master/pkg/specs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Task scheduled Test", func() {

	Context("Order Tasks", func() {

		It("Order1", func() {

			unorderedTasks := []TaskScheduled{
				TaskScheduled{
					Task: &Task{
						Name:     "Task1",
						Priority: 10,
					},
				},
				TaskScheduled{
					Task: &Task{
						Name:     "Task2",
						Priority: 14,
					},
				},
			}

			sort.Sort(TaskSchedPrioritySorter(unorderedTasks))

			Expect(unorderedTasks[0].Task.Name).To(Equal("Task1"))
			Expect(unorderedTasks[1].Task.Name).To(Equal("Task2"))
		})

		It("Order2", func() {

			unorderedTasks := []TaskScheduled{
				TaskScheduled{
					Task: &Task{
						Name:     "Task1",
						Priority: 15,
					},
				},
				TaskScheduled{
					Task: &Task{
						Name:     "Task2",
						Priority: 14,
					},
				},
				TaskScheduled{
					Task: &Task{
						Name:     "Task3",
						Priority: 1,
					},
				},
			}

			sort.Sort(TaskSchedPrioritySorter(unorderedTasks))

			Expect(unorderedTasks[0].Task.Name).To(Equal("Task3"))
			Expect(unorderedTasks[1].Task.Name).To(Equal("Task2"))
			Expect(unorderedTasks[2].Task.Name).To(Equal("Task1"))
		})

	})

})
