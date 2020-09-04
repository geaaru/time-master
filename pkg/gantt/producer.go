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
package gantt

import (
	"errors"

	log "github.com/geaaru/time-master/pkg/logger"
	specs "github.com/geaaru/time-master/pkg/specs"
)

type TimeMasterGanttProducer interface {
	Build(*specs.ScenarioSchedule, ProducerOpts) ([]byte, error)
}

type DefaultGanttProducer struct {
	Logger *log.TmLogger
	Config *specs.TimeMasterConfig
}

type ProducerOpts struct {
	ShowActivityOnTasks bool
	OrderByEndTime      bool
}

func NewProducer(config *specs.TimeMasterConfig, t string) (TimeMasterGanttProducer, error) {
	var ans TimeMasterGanttProducer
	switch t {
	case "frappe":
		ans = NewFrappeGanttProducer(config)
	default:
		return ans, errors.New("Invalid producer type")
	}

	return ans, nil
}

func newDefaultGanttProducer(config *specs.TimeMasterConfig) *DefaultGanttProducer {
	ans := &DefaultGanttProducer{
		Config: config,
		Logger: log.NewTmLogger(config),
	}

	// Initialize logging
	if config.GetLogging().EnableLogFile && config.GetLogging().Path != "" {
		err := ans.Logger.InitLogger2File()
		if err != nil {
			ans.Logger.Fatal("Error on initialize logfile")
		}
	}

	return ans
}
