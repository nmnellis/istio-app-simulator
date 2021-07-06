package cmd

import (
	"fmt"
	"github.com/nmnellis/app-gen/pkg/generate"
	"github.com/spf13/cobra"
	"os"
)

var cfg = &generate.Config{}

var generateCmd = &cobra.Command{
	Use:     "generate",
	Aliases: []string{"g", "gen"},
	Short:   "generate kubernetes yaml for applications",
	Long:    `This application generates kubernetes yaml that sets up complex application networks for a given amount of namespaces. `,
	RunE: func(cmd *cobra.Command, args []string) error {
		return generate.NewAppGenerator(cfg).Generate()
	},
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func Execute() {
	// used to generate docs
	// err := doc.GenMarkdownTree(generateCmd, "docs")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	if err := generateCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	generateCmd.Flags().StringVar(&cfg.Hostname, "hostname", "*", "Hostname to use for gateway and virtualService")
	generateCmd.Flags().Int64Var(&cfg.Seed, "seed", 0, "Override random seed with static one (for deterministic outputs)")
	generateCmd.Flags().IntVarP(&cfg.NumberOfNamespaces, "namespaces", "n", 1, "Number of namespaces to generate applications for")
	generateCmd.Flags().IntVarP(&cfg.NumberOfTiers, "tiers", "t", 3, "Length of the application call stack per namespace ("+
		"how many applications deep)")
	generateCmd.Flags().IntVar(&cfg.MaxAppsPerTier, "apps-per-tier", 5, "Max amount of applications that can exist in a given tier. "+
		"Will randomly pick between 1 < x")
	generateCmd.Flags().IntVar(&cfg.ChanceOfVersions, "chance-version", 10,
		"Percent chance that a given application will have multiple versions v1/v2/v3 (0-100)")
	generateCmd.Flags().IntVar(&cfg.ChanceOfCrossNamespaceChatter, "chance-cross-namespace-chatter", 10,
		"Percent chance that a given application will make a call to an application in another namespace (0-100)")
	generateCmd.Flags().IntVar(&cfg.ChanceOfErrors, "chance-of-app-errors", 0,
		"Percent chance that a given application will return errors in their responses (0-100)")
	generateCmd.Flags().Float32Var(&cfg.ErrorPercent, "app-error-percent", .05,
		"If an application returns errors, what percent of requests should have errors? (0-1)")
	generateCmd.Flags().IntVar(&cfg.ChanceToCallExternalService, "chance-call-external", 10,
		"Percent chance that a given application will make a call to an external service (0-100)")
	generateCmd.Flags().StringVarP(&cfg.OutputDir, "output-dir", "o", "out",
		"Output directory where assets will be generated")

	generateCmd.Flags().StringVar(&cfg.MemoryRequest, "requests-memory", "100Mi",
		"Kubernetes container memory request")
	generateCmd.Flags().StringVar(&cfg.MemoryLimit, "limits-memory", "",
		"Kubernetes container memory limit")
	generateCmd.Flags().StringVar(&cfg.CPULimit, "limits-cpu", "",
		"Kubernetes container CPU limit")
	generateCmd.Flags().StringVar(&cfg.CPURequest, "requests-cpu", "100m",
		"Kubernetes container CPU request")
}
