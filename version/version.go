package version

import (
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Name      = ""
	Version   = ""
	GitCommit = ""
	BuildDate = ""
	BuildTags = ""
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the application binary version information",
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Printf("Name: %s\n", Name)
			cmd.Printf("Version: %s\n", Version)
			cmd.Printf("Git Commit: %s\n", GitCommit)
			cmd.Printf("Build Date: %s\n", BuildDate)
			cmd.Printf("Build Tags: %s\n", BuildTags)
			cmd.Printf("Go Version: %s\n", runtime.Version())
			cmd.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
	return cmd
}
