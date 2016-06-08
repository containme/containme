package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version of containme
var Version = "0.1.0+git"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of containme",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
