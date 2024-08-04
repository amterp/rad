package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"rad/core"
)

var rootCmd = &cobra.Command{
	Use:   "rad",
	Short: "Request And Display (RAD)",
	Long:  `Request And Display (RAD): A tool for making HTTP requests, extracting details, and displaying the result.`,
	Args:  cobra.MinimumNArgs(1), // todo 0 maybe should just display help w/out error?
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	Run: func(cmd *cobra.Command, args []string) {
		script := args[0]
		source := readSource(&script)
		_ = core.NewLexer(source)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func readSource(script *string) *string {
	source, err := os.ReadFile(*script)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read script '%s': %v\n", *script, err)
		os.Exit(1)
	}
	sourceStr := string(source)
	return &sourceStr
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
}
