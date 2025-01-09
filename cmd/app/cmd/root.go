package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "HOKI Tabloid Backend Server Application",
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().String("env", "", "Which environment this server will run on")

}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
