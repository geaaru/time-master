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
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"

	specs "github.com/geaaru/time-master/pkg/specs"

	helpers "github.com/mudler/luet/pkg/helpers"
)

func (i *TimeMasterInstance) Load() error {

	// Load clients
	for _, dir := range i.Config.GetClientsDirs() {
		// Ignore error on load directory
		i.LoadClientDir(dir)
	}

	// Load resources
	for _, dir := range i.Config.GetResourcesDirs() {
		// Ignore error on load directory
		i.LoadResourceDir(dir)
	}

	// Load timesheets
	for _, dir := range i.Config.GetTimesheetsDirs() {
		// Ignore error on load directory
		i.LoadTimesheetsDir(dir)
	}

	// Load scenarios
	for _, dir := range i.Config.GetScenariosDirs() {
		// Ignore error on load directory
		i.LoadScenarioDir(dir)
	}

	return nil
}

func (i *TimeMasterInstance) loadExtraClientFiles(client *specs.Client) error {
	clientBaseDir, err := filepath.Abs(path.Dir(client.File))
	if err != nil {
		return err
	}

	if len(client.ActivitiesDirs) > 0 {
		for _, adir := range client.ActivitiesDirs {

			dir := path.Join(clientBaseDir, adir)

			if !helpers.Exists(dir) {
				i.Logger.Debug("For client", client.Name, "activity dir", adir,
					"is not present.")
				continue
			}

			err := i.LoadClientActivities(client.Name, dir)
			if err != nil {
				return err
			}

		}
	}

	return nil
}

func (i *TimeMasterInstance) LoadScenarioDir(dir string) error {
	var regexConfs = regexp.MustCompile(`.yml$|.yaml$`)

	i.Logger.Debug("Checking directory", dir, "...")

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		i.Logger.Debug("Skip dir", dir, ":", err.Error())
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !regexConfs.MatchString(file.Name()) {
			i.Logger.Debug("File", file.Name(), "skipped.")
			continue
		}

		content, err := ioutil.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			i.Logger.Warning("On read file", file.Name(), ":", err.Error())
			i.Logger.Debug("File", file.Name(), "skipped.")
			continue
		}

		scenario, err := specs.ScenarioFromYaml(content, path.Join(dir, file.Name()))
		if err != nil {
			i.Logger.Warning("On parse file", file.Name(), ":", err.Error())
			i.Logger.Debug("File", file.Name(), "skipped.")
			continue
		}

		i.AddScenario(scenario)

		i.Logger.Debug("Loaded scenario", scenario.Name, ".")
	}

	return nil
}

func (i *TimeMasterInstance) LoadResourceDir(dir string) error {
	var regexConfs = regexp.MustCompile(`.yml$|.yaml$`)

	i.Logger.Debug("Checking directory", dir, "...")

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		i.Logger.Debug("Skip dir", dir, ":", err.Error())
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !regexConfs.MatchString(file.Name()) {
			i.Logger.Debug("File", file.Name(), "skipped.")
			continue
		}

		content, err := ioutil.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			i.Logger.Debug("On read file", file.Name(), ":", err.Error())
			i.Logger.Debug("File", file.Name(), "skipped.")
			continue
		}

		resource, err := specs.ResourceFromYaml(content, path.Join(dir, file.Name()))
		if err != nil {
			i.Logger.Debug("On parse file", file.Name(), ":", err.Error())
			i.Logger.Debug("File", file.Name(), "skipped.")
			continue
		}

		i.AddResource(resource)

		i.Logger.Debug(fmt.Sprintf(
			"Loaded resource %20s - %s :check_mark:", resource.User, resource.Name))
	}

	return nil
}

func (i *TimeMasterInstance) LoadClientDir(dir string) error {
	var regexConfs = regexp.MustCompile(`.yml$|.yaml$`)

	i.Logger.Debug("Checking directory", dir, "...")

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		i.Logger.Debug("Skip dir", dir, ":", err.Error())
		return err
	}

	for _, file := range files {

		if file.IsDir() {
			continue
		}

		if !regexConfs.MatchString(file.Name()) {
			i.Logger.Debug("File", file.Name(), "skipped.")
			continue
		}

		content, err := ioutil.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			i.Logger.Debug("On read file", file.Name(), ":", err.Error())
			i.Logger.Debug("File", file.Name(), "skipped.")
			continue
		}

		client, err := specs.ClientFromYaml(content, path.Join(dir, file.Name()))
		if err != nil {
			i.Logger.Debug("On parse file", file.Name(), ":", err.Error())
			i.Logger.Debug("File", file.Name(), "skipped.")
			continue
		}

		i.AddClient(client)

		err = i.loadExtraClientFiles(client)
		if err != nil {
			i.Logger.Debug("Error on load client extra files: " + err.Error())
			return err
		}

		i.Logger.Debug("Loaded client " + client.Name)

	}

	return nil
}

func (i *TimeMasterInstance) LoadClientActivities(clientName, dir string) error {
	var regexConfs = regexp.MustCompile(`.yml$|.yaml$`)

	client, err := i.GetClientByName(clientName)
	if err != nil {
		return err
	}

	// Load activities
	i.Logger.Debug("Checking directory", dir, "...")

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		i.Logger.Debug("Skip dir", dir, ":", err.Error())
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !regexConfs.MatchString(file.Name()) {
			i.Logger.Debug("File", file.Name(), "skipped.")
			continue
		}

		content, err := ioutil.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			i.Logger.Debug("On read file", file.Name(), ":", err.Error())
			i.Logger.Debug("File", file.Name(), "skipped.")
			continue
		}

		activity, err := specs.ActivityFromYaml(content, path.Join(dir, file.Name()))
		if err != nil {
			i.Logger.Warning("On parse file", file.Name(), ":", err.Error())
			i.Logger.Debug("File", file.Name(), "skipped.")
			continue
		}

		if activity.Disabled {
			i.Logger.Debug("Skipping disabled acivity " + activity.Name)
		} else {
			i.Logger.Debug("Loading activity " + activity.Name)

			activity.InitDefaultPriority(i.Config.GetWork().TaskDefaultPriority)

			client.AddActivity(*activity)
		}

	}

	return nil
}

func (i *TimeMasterInstance) LoadTimesheetsDir(dir string) error {
	var regexConfs = regexp.MustCompile(`.yml$|.yaml$`)

	i.Logger.Debug("Checking directory", dir, "...")

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		i.Logger.Debug("Skip dir", dir, ":", err.Error())
		return err
	}

	for _, file := range files {

		if file.IsDir() {
			continue
		}

		if !regexConfs.MatchString(file.Name()) {
			i.Logger.Debug("File", file.Name(), "skipped.")
			continue
		}

		content, err := ioutil.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			i.Logger.Debug("On read file", file.Name(), ":", err.Error())
			i.Logger.Debug("File", file.Name(), "skipped.")
			continue
		}

		agenda, err := specs.AgengaTimesheetFromYaml(content, path.Join(dir, file.Name()))
		if err != nil {
			i.Logger.Debug("On parse file", file.Name(), ":", err.Error())
			i.Logger.Debug("File", file.Name(), "skipped.")
			continue
		}

		i.AddAgendaTimesheet(agenda)

		if agenda.Name != "" {
			i.Logger.Debug(fmt.Sprintf("Loaded agenda %s with %d timesheets.",
				agenda.Name, len(agenda.Timesheets)))
		} else {
			i.Logger.Debug(fmt.Sprintf("Loaded agenda %s with %d timesheets.",
				path.Base(agenda.File), len(agenda.Timesheets)))
		}

	}

	return nil
}
