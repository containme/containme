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
	"path/filepath"

	"github.com/containme/containme/builder"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var noCache bool

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run a build",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		workdir, err = filepath.Abs(workdir)
		if err != nil {
			return errors.Wrapf(err, "could not find workdir")
		}

		buildSpec, err := builder.ParseBuildSpecFile(cfgFile)
		if err != nil {
			return errors.Wrapf(err, "failed to parse %s", cfgFile)
		}

		profileResolver, err := builder.NewProfileResolver(buildSpec.Environment.Profile)
		if err != nil {
			return errors.Wrapf(err, "failed to resolve profile")
		}

		profile, err := profileResolver.ResolveProfile()
		if err != nil {
			return errors.Wrapf(err, "failed to resolve profile")
		}

		b, err := builder.NewBuilder(workdir, buildSpec, profile)
		if err != nil {
			return errors.Wrapf(err, "failed to create builder")
		}
		b.DebugThisShit()

		cacheID, err := b.ExecuteEnivronmentStage(!noCache, state.EnvironmentImageCache)
		if err != nil {
			return errors.Wrapf(err, "failed to execute environment stage")
		}
		if cacheID != "" {
			state.EnvironmentImageCache = cacheID
			WriteState()
		}

		err = b.ExecuteDependenciesStage(false)
		if err != nil {
			return errors.Wrapf(err, "failed to execute dependencies stage")
		}

		err = b.ExecuteTestStage(false)
		if err != nil {
			return errors.Wrapf(err, "failed to execute test stage")
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(runCmd)

	runCmd.Flags().BoolVarP(&noCache, "no-cache", "N", false, "disable environment cache for this build.")

}
