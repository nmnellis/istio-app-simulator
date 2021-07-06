package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "app-gen",
	Short: "Test Application Network Generator",
	Long:  `Generates a set of applications that communicate with each other in a defined network`,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func Execute() {
	// used to generate docs
	// err := doc.GenMarkdownTree(rootCmd, "docs")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
