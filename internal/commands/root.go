package commands

import (
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var rootArgs = struct {
	Verbose bool
}{
	Verbose: false,
}

var rootCmd = &cobra.Command{
	Use:           "cf-ddns",
	SilenceUsage:  true,
	SilenceErrors: true,
	Short:         `Dynamic DNS service based on Cloudflare`,
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		if rootArgs.Verbose {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	flags := rootCmd.PersistentFlags()
	flags.BoolVar(&rootArgs.Verbose, "verbose", rootArgs.Verbose, "be verbose")

	rootCmd.AddCommand(
		watchCmd,
	)
}
