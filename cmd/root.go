// Copyright © 2017 Yeho Nazarkin <nimnull@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go_scoring/validate"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "go_scoring <project_id> <model_id> <dataset path>",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
// Uncomment the following line if your bare application
// has an action associated with it:
	PreRun: func(cmd *cobra.Command, args []string) {
		hostURL = validate.ValidateHost(
			viper.GetString("host"))
		validate.ValidateFile(viper.GetString("out"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Root called")
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.go_scoring.yaml)")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.PersistentFlags().String("host", "",
		"server FQDN prefixed with a protocol type (http://server.domain)")
	RootCmd.PersistentFlags().String("out", "out.csv",
		`Specifies the file name and optionally path, to which the results are written.
		If not specified, the default file name is out.csv, written to the directory containing
		the script.`)
	RootCmd.PersistentFlags().String("encoding", "",
		`Specifies dataset encoding. If not provided, the script attempts to
		detect the decoding (e.g., "utf-8", "latin-1", or "iso2022_jp").`)
	RootCmd.PersistentFlags().String("delimiter", "",
		`Specifies the delimiter to recognize in the input .csv file. E.g.
		"--delimiter=,". If not specified, the script tries to automatically determine
		the delimiter. The special keyword "tab" can be used to indicate a tab delimited
		csv. "pipe" can be used to indicate "|"`)
	viper.BindPFlag("host", RootCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("out", RootCmd.PersistentFlags().Lookup("out"))
	viper.BindPFlag("encoding", RootCmd.PersistentFlags().Lookup("encoding"))
	viper.BindPFlag("delimiter", RootCmd.PersistentFlags().Lookup("delimiter"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".go_scoring") // name of config file (without extension)
	viper.AddConfigPath("$HOME")  // adding home directory as first search path
	viper.AutomaticEnv()          // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
