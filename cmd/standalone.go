// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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
	"log"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go_scoring/scoring"
	"go_scoring/validate"
)

var hostURL *url.URL

// standaloneCmd represents the standalone command
var standaloneCmd = &cobra.Command{
	Use:   "standalone",
	Short: "Scoring over standalone prediction instances",
	Long:  `Use: go_scoring standalone [flags] <import_id> <dataset path>`,
	PreRun: func(cmd *cobra.Command, args []string) {
		hostURL = validate.ValidateHost(
			viper.GetString("host"))
		validate.ValidateFile(viper.GetString("out"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			log.Fatal("Not all necessary arguments provided")
		}
		importId := args[0]
		dataset := args[1]
		// TODO: Work your own magic here
		scoring.RunBatch(
			hostURL, importId, dataset,
			viper.GetString("encoding"),
			viper.GetString("delimiter"),
			0, 0, false, false)
	},
}

func init() {
	RootCmd.AddCommand(standaloneCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// standaloneCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// standaloneCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
