//go:build full || build

package subcommands

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"lunchpail.io/cmd/options"
	"lunchpail.io/pkg/fe/builder"
	"lunchpail.io/pkg/fe/builder/overlay"
)

func newBuildCmd() *cobra.Command {
	var outputFlag string
	var branchFlag string
	var sourceIsYaml bool
	var allFlag bool

	cmd := &cobra.Command{
		Use:     "build [path-or-git]",
		GroupID: applicationGroup.ID,
		Short:   "Generate a binary specialized to a given application",
		Long:    "Generate a binary specialized to a given application",
		Args:    cobra.MatchAll(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
	}

	cmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Path to store output binary")
	if err := cmd.MarkFlagRequired("output"); err != nil {
		log.Fatalf("Required option -o/--output <outputPath>")
	}

	cmd.Flags().StringVarP(&branchFlag, "branch", "b", branchFlag, "Git branch to pull from")
	cmd.Flags().BoolVarP(&sourceIsYaml, "yaml", "y", sourceIsYaml, "The source directory contains high-level IR YAML")
	cmd.Flags().BoolVarP(&allFlag, "all-platforms", "A", allFlag, "Generate binaries for all supported platform/arch combinations")

	var command string
	cmd.Flags().StringVarP(&command, "command", "c", command, "Run the given program given as a string")

	buildOptions, err := options.AddBuildOptions(cmd)
	if err != nil {
		panic(err)
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		sourcePath := ""
		if len(args) >= 1 {
			sourcePath = args[0]
		}

		overrideValues, err := cmd.Flags().GetStringSlice("set")
		if err != nil {
			return err
		} else {
			buildOptions.OverrideValues = overrideValues
		}

		overrideFileValues, err := cmd.Flags().GetStringSlice("set-file")
		if err != nil {
			return err
		} else {
			buildOptions.OverrideFileValues = overrideFileValues
		}

		return builder.Build(context.Background(), sourcePath, builder.Options{
			Name:         outputFlag,
			AllPlatforms: allFlag,
			OverlayOptions: overlay.Options{
				Branch:       branchFlag,
				Command:      command,
				SourceIsYaml: sourceIsYaml,
				BuildOptions: *buildOptions,
			},
		})
	}

	return cmd
}

func init() {
	rootCmd.AddCommand(newBuildCmd())
}
