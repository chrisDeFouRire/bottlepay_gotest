package cmd

import (
	"github.com/bottlepay/portfolio-data/store"
	"github.com/spf13/cobra"
)

// trackCmd represents the track command
var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "Start the tracker service",
	Long:  `The tracker service tracks the portfolios of users`,
	RunE: func(cmd *cobra.Command, args []string) error {

		s := store.NewFakeUserStore()
		s.Populate()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(trackCmd)
}
