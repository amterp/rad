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
		scriptPath := args[0]
		source := readSource(scriptPath)
		l := core.NewLexer(source)
		l.Lex()
		fmt.Println()
		for _, token := range l.Tokens {
			fmt.Println(token)
			fmt.Println()
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func readSource(scriptPath string) string {
	source, err := os.ReadFile(scriptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read script '%s': %v\n", scriptPath, err)
		os.Exit(1)
	}
	return string(source)
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
}
