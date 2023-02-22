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
package importer

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/geaaru/time-master/pkg/logger"
	specs "github.com/geaaru/time-master/pkg/specs"

	"gopkg.in/yaml.v2"
)

type ImportOpts struct {
	SplitResource bool
}

type TimeMasterImporter interface {
	LoadTimesheets(string) error
	WriteTimesheets() error
	GetTimesheets() *[]specs.AgendaTimesheets
	AddTimesheet(*specs.AgendaTimesheets)
}

type DefaultImporter struct {
	Logger       *log.TmLogger
	Config       *specs.TimeMasterConfig
	TimesheetDir string
	FilePrefix   string
	Opts         ImportOpts
	Timesheets   []specs.AgendaTimesheets
}

func NewDefaultImporter(config *specs.TimeMasterConfig, tmDir, filePrefix string, opts ImportOpts) *DefaultImporter {
	ans := &DefaultImporter{
		Config:       config,
		Logger:       log.NewTmLogger(config),
		TimesheetDir: tmDir,
		FilePrefix:   filePrefix,
		Opts:         opts,
		Timesheets:   []specs.AgendaTimesheets{},
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

func (i *DefaultImporter) GetTimesheets() *[]specs.AgendaTimesheets {
	return &i.Timesheets
}

func (i *DefaultImporter) AddTimesheet(t *specs.AgendaTimesheets) {
	i.Timesheets = append(i.Timesheets, *t)
}

func (i *DefaultImporter) WriteTimesheets() error {
	// Ensure that timesheetDir is an absolute path to avoid errors with path.Jon
	tmDir, err := filepath.Abs(i.TimesheetDir)
	if err != nil {
		return err
	}

	for _, agenda := range i.Timesheets {
		if i.FilePrefix == "" && agenda.File == "" {
			return errors.New("Both file prefix and agenda file are empty")
		}

		tmFile := filepath.Join(tmDir, fmt.Sprintf("%s%s.yml", i.FilePrefix, agenda.File))

		data, err := yaml.Marshal(agenda)
		if err != nil {
			return err
		}

		dirName := filepath.Dir(tmFile)
		if _, serr := os.Stat(dirName); serr != nil {
			err = os.MkdirAll(dirName, os.ModePerm)
			if err != nil {
				return err
			}
		}

		err = ioutil.WriteFile(tmFile, data, 0644)
		if err != nil {
			return err
		}

		i.Logger.Info(fmt.Sprintf(">>> [timesheet] Created file %s :check_mark:", tmFile))

	}

	return nil
}
