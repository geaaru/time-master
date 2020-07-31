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
	specs "github.com/geaaru/time-master/pkg/specs"
)

type TimeMasterInstance struct {
	Logger    *log.TmLogger
	Config    *specs.TimeMasterConfig
	Clients   []specs.Client
	Resources []specs.Resource
	Scenarios []specs.Scenario
}

func NewTimeMasterInstance(config *specs.TimeMasterConfig) *TimeMasterInstance {
	ans := &TimeMasterInstance{
		Config:    config,
		Logger:    log.NewTmLogger(config),
		Clients:   []specs.Client{},
		Resources: []specs.Resource{},
		Scenarios: []specs.Scenario{},
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

func (i *TimeMasterInstance) AddScenario(s *specs.Scenario) {
	i.Scenarios = append(i.Scenarios, *s)
}

func (i *TimeMasterInstance) GetScenarios() *[]specs.Scenario {
	return &i.Scenarios
}

func (i *TimeMasterInstance) GetClientByName(name string) (specs.Client, error) {
	ans := specs.Client{}

	for _, c := range i.Clients {
		if c.Name == name {
			return c, nil
		}
	}
	return ans, errors.New("Client " + name + " not present")
}
