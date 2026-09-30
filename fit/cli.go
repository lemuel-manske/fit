package fit

import (
	"fmt"
	"os"
	"strings"

	fit_cmd "fit/fit/cmd"
	"fit/fit/internal"

	"path/filepath"

	"github.com/spf13/cobra"
)

func InitCLI(root *cobra.Command) {
	root.AddCommand(NewAddCmd())
	root.AddCommand(NewCheckoutCmd())
	root.AddCommand(NewCommitCmd())
	root.AddCommand(NewInitCmd())
	root.AddCommand(NewMergeCmd())
	root.AddCommand(NewRmCmd())
}

func Execute(cmd *cobra.Command) {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use: "fit",
	}
}

func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use: "init",

		Args: cobra.ExactArgs(0),

		RunE: func(cmd *cobra.Command, args []string) error {
			currDir, err := os.Getwd()
			if err != nil {
				return err
			}

			repositoryName := strings.TrimSuffix(filepath.Base(currDir), filepath.Ext(currDir))

			return fit_cmd.Init(currDir, repositoryName)
		},
	}
}

func NewAddCmd() *cobra.Command {
	return &cobra.Command{
		Use: "add",

		Args: cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			currDir, err := os.Getwd()
			if err != nil {
				return err
			}

			return fit_cmd.Add(currDir, args[0])
		},
	}
}

func NewRmCmd() *cobra.Command {
	return &cobra.Command{
		Use: "rm",

		Args: cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			currDir, err := os.Getwd()
			if err != nil {
				return err
			}

			return fit_cmd.Rm(currDir, args[0])
		},
	}
}

func NewCommitCmd() *cobra.Command {
	return &cobra.Command{
		Use: "commit",

		Args: cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			currDir, err := os.Getwd()
			if err != nil {
				return err
			}

			commitID, err := fit_cmd.CommitChanges(currDir, args[0])

			if err != nil {
				return err
			}

			fmt.Println(commitID)
			return nil
		},
	}
}

func NewMergeCmd() *cobra.Command {
	return &cobra.Command{
		Use: "merge",

		Args: cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			currDir, err := os.Getwd()
			if err != nil {
				return err
			}

			return fit_cmd.Merge(currDir, internal.CommitID(args[0]))
		},
	}
}

func NewCheckoutCmd() *cobra.Command {
	return &cobra.Command{
		Use: "checkout",

		Args: cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			currDir, err := os.Getwd()
			if err != nil {
				return err
			}

			return fit_cmd.Checkout(currDir, internal.CommitID(args[0]))
		},
	}
}
