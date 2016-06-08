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

var noEnvCache bool
var noDepsCache bool

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run a build",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		workdir, err = filepath.Abs(workdir)
		if err != nil {
			ExitWithError(ExitError, errors.Wrapf(err, "could not find workdir"))
		}

		buildSpec, err := builder.ParseBuildSpecFile(cfgFile)
		if err != nil {
			ExitWithError(ExitError, errors.Wrapf(err, "failed to parse %s", cfgFile))
		}

		profileResolver, err := builder.NewProfileResolver(buildSpec.Environment.Profile)
		if err != nil {
			ExitWithError(ExitError, errors.Wrapf(err, "failed to resolve profile"))
		}

		profile, err := profileResolver.ResolveProfile()
		if err != nil {
			ExitWithError(ExitError, errors.Wrapf(err, "failed to resolve profile"))
		}

		b, err := builder.NewBuilder(workdir, buildSpec, profile)
		if err != nil {
			ExitWithError(ExitError, errors.Wrapf(err, "failed to create builder"))
		}

		cacheID, err := b.ExecuteEnivronmentStage(!noEnvCache, state.EnvironmentImageCache)
		if err != nil {
			ExitWithError(ExitError, errors.Wrapf(err, "failed to execute environment stage"))
		}
		if cacheID != "" {
			state.EnvironmentImageCache = cacheID
			WriteState()
		}

		cacheID, err = b.ExecuteDependenciesStage(!noEnvCache && !noDepsCache, state.DependenciesImageCache)
		if cacheID != "" {
			state.DependenciesImageCache = cacheID
			WriteState()
		}
		if err != nil {
			ExitWithError(ExitError, errors.Wrapf(err, "failed to execute dependencies stage"))
		}

		err = b.ExecuteTestStage()
		if err != nil {
			ExitWithError(ExitError, errors.Wrapf(err, "failed to execute test stage"))
		}
	},
}

func init() {
	RootCmd.AddCommand(runCmd)

	runCmd.Flags().BoolVarP(&noEnvCache, "no-environment-cache", "E", false, "disable environment cache for this build.")
	runCmd.Flags().BoolVarP(&noDepsCache, "no-dependencies-cache", "D", false, "disable dependencies cache for this build.")

}
