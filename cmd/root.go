// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
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
	"os"
	"path"

	"github.com/spf13/cobra"
)

var cfgFile string
var workdir string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "containme",
	Short: "contain your builds",
	Long:  ``,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVarP(&cfgFile, "containme-file", "f", "./containme.yaml", "containme file to use for builds")
	RootCmd.PersistentFlags().StringVarP(&workdir, "workdir", "w", "./", "working directory for build")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	os.Mkdir(path.Join(workdir, ".containme"), 0755)
	LoadState()
}
