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
package loader

import (
	"errors"

	log "github.com/geaaru/time-master/pkg/logger"
	scheduler "github.com/geaaru/time-master/pkg/scheduler"
	specs "github.com/geaaru/time-master/pkg/specs"
)

type TimeMasterInstance struct {
	Logger     *log.TmLogger
	Config     *specs.TimeMasterConfig
	Clients    []specs.Client
	Resources  []specs.Resource
	Scenarios  []specs.Scenario
	Timesheets []specs.AgendaTimesheets
}

func NewTimeMasterInstance(config *specs.TimeMasterConfig) *TimeMasterInstance {
	ans := &TimeMasterInstance{
		Config:     config,
		Logger:     log.NewTmLogger(config),
		Clients:    []specs.Client{},
		Resources:  []specs.Resource{},
		Scenarios:  []specs.Scenario{},
		Timesheets: []specs.AgendaTimesheets{},
	}

	// Initialize logging
	if config.GetLogging().EnableLogFile && config.GetLogging().Path != "" {
		err := ans.Logger.InitLogger2File()
		if err != nil {
			ans.Logger.Fatal("Error on initialize logfile")
		}
	}
	ans.Logger.SetAsDefault()

	return ans
}

func (i *TimeMasterInstance) AddClient(c *specs.Client) {
	i.Clients = append(i.Clients, *c)
}

func (i *TimeMasterInstance) GetClients() *[]specs.Client {
	return &i.Clients
}

func (i *TimeMasterInstance) AddResource(r *specs.Resource) {
	i.Resources = append(i.Resources, *r)
}

func (i *TimeMasterInstance) GetResources() *[]specs.Resource {
	return &i.Resources
}

func (i *TimeMasterInstance) GetResourceByUser(user string) *specs.Resource {
	for idx, r := range i.Resources {
		if r.User == user {
			return &i.Resources[idx]
		}
	}
	return nil
}

func (i *TimeMasterInstance) AddScenario(s *specs.Scenario) {
	i.Scenarios = append(i.Scenarios, *s)
}

func (i *TimeMasterInstance) GetScenarios() *[]specs.Scenario {
	return &i.Scenarios
}

func (i *TimeMasterInstance) GetTimesheets() *[]specs.AgendaTimesheets {
	return &i.Timesheets
}

func (i *TimeMasterInstance) AddAgendaTimesheet(t *specs.AgendaTimesheets) {
	i.Timesheets = append(i.Timesheets, *t)
}

func (i *TimeMasterInstance) GetClientByName(name string) (*specs.Client, error) {
	for idx, c := range i.Clients {
		if c.Name == name {
			return &i.Clients[idx], nil
		}
	}
	return nil, errors.New("Client " + name + " not present")
}

func (i *TimeMasterInstance) GetScenarioByName(name string) (*specs.Scenario, error) {
	for idx, s := range i.Scenarios {
		if s.Name == name {
			return &i.Scenarios[idx], nil
		}
	}
	return nil, errors.New("Scenario " + name + " not present")
}

func (i *TimeMasterInstance) InitScheduler(sched scheduler.TimeMasterScheduler) {
	sched.SetClients(i.GetClients())
	sched.SetResources(i.GetResources())
	sched.SetTimesheets(i.GetTimesheets())
}

func (i *TimeMasterInstance) GetAllTaskMap() map[string]specs.Task {
	ans := make(map[string]specs.Task, 0)

	// Retrieve the list of all tasks
	for _, client := range i.Clients {
		for _, activity := range *client.GetActivities() {
			aTasks := activity.GetAllTasksList()
			for _, t := range aTasks {
				ans[t.Name] = t
			}
		}
	}

	return ans
}
