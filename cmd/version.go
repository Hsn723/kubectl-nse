package cmd

import (
	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
)

var (
	// CurrentVersion holds the version number of kubectl-nse, set at build time
	CurrentVersion string
	versionCmd     = &cobra.Command{
		Use:   "version",
		Short: "show the version number",
		Long:  "Show the version number for kubectl-nse and checks if it currently is the latest version.string",
		Args:  cobra.NoArgs,
		RunE:  runVersion,
	}
)

func runVersion(cmd *cobra.Command, args []string) (err error) {
	currentVersion, err := version.NewVersion(CurrentVersion)
	if err != nil {
		return
	}
	cmd.Printf("kubectl-nse version %s\n", currentVersion)
	return
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
