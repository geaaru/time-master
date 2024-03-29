/*
Copyright (C) 2020-2021  Daniele Rondina <geaaru@funtoo.org>
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
package cmd

import (
	"fmt"
	"os"
	"strings"

	specs "github.com/geaaru/time-master/pkg/specs"
	utils "github.com/geaaru/time-master/pkg/tools"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	cliName = `Copyright (c) 2020-2024 - Daniele Rondina

Time Master`

	TM_VERSION = `0.5.0`
)

var (
	BuildTime   string
	BuildCommit string
)

func initConfig(config *specs.TimeMasterConfig) {
	// Set env variable
	config.Viper.SetEnvPrefix(specs.TM_ENV_PREFIX)
	config.Viper.BindEnv("config")
	config.Viper.SetDefault("config", "")
	config.Viper.SetDefault("etcd-config", false)

	config.Viper.AutomaticEnv()

	// Create EnvKey Replacer for handle complex structure
	replacer := strings.NewReplacer(".", "__")
	config.Viper.SetEnvKeyReplacer(replacer)

	// Set config file name (without extension)
	config.Viper.SetConfigName(specs.TM_CONFIGNAME)

	config.Viper.SetTypeByDefaultValue(true)

}

func initCommand(rootCmd *cobra.Command, config *specs.TimeMasterConfig) {
	var pflags = rootCmd.PersistentFlags()

	pflags.StringP("config", "c", "", "Time Master configuration file")
	pflags.BoolP("verbose", "v", config.Viper.GetBool("general.debug"), "Verbose output.")

	config.Viper.BindPFlag("config", pflags.Lookup("config"))
	config.Viper.BindPFlag("general.debug", pflags.Lookup("verbose"))

	rootCmd.AddCommand(
		newClientCommand(config),
		newActivityCommand(config),
		newChangeRequestCommand(config),
		newPrintCommand(config),
		newValidateCommand(config),
		newImportCommand(config),
		newResourceCommand(config),
		newTimesheetCommand(config),
		newTaskCommand(config),
		newScenarioCommand(config),
		newGanttCommand(config),
	)
}

func Execute() {
	// Create Main Instance Config object
	var config *specs.TimeMasterConfig = specs.NewTimeMasterConfig(nil)

	initConfig(config)

	var rootCmd = &cobra.Command{
		Short:        cliName,
		Version:      fmt.Sprintf("%s-g%s %s", TM_VERSION, BuildCommit, BuildTime),
		Args:         cobra.OnlyValidArgs,
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				os.Exit(0)
			}
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var err error
			var v *viper.Viper = config.Viper

			v.SetConfigType("yml")
			if v.Get("config") == "" {
				config.Viper.AddConfigPath(".")
			} else {
				v.SetConfigFile(v.Get("config").(string))
			}

			// Parse configuration file
			err = config.Unmarshal()
			utils.CheckError(err)
		},
	}

	initCommand(rootCmd, config)

	// Start command execution
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
