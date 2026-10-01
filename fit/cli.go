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

		Short: "Initializes a new fit repository",

		Args: cobra.ExactArgs(0),

		RunE: func(cmd *cobra.Command, args []string) error {
			currDir, err := os.Getwd()
			if err != nil {
				return err
			}

			repositoryName := strings.TrimSuffix(
				filepath.Base(currDir),
				filepath.Ext(currDir),
			)

			return fit_cmd.Init(currDir, repositoryName)
		},
	}
}

func NewAddCmd() *cobra.Command {
	return &cobra.Command{
		Use: "add",

		Short: "Add a file or directory to the index",

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

		Short: "Remove a file from the index",

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
	var message string

	cmd := &cobra.Command{
		Use: "commit",

		Short: "Commits the current changes in the index",

		Args: cobra.ExactArgs(0),

		RunE: func(cmd *cobra.Command, args []string) error {
			currDir, err := os.Getwd()
			if err != nil {
				return err
			}

			commitID, err := fit_cmd.CommitChanges(currDir, message)

			if err != nil {
				return err
			}

			fmt.Println(commitID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Commit message")

	return cmd
}

func NewMergeCmd() *cobra.Command {
	var abort bool

	cmd := &cobra.Command{
		Use: "merge [commit]",

		Short: "Merges a specific commit into HEAD",

		Args: func(cmd *cobra.Command, args []string) error {
			if abort {
				return cobra.NoArgs(cmd, args)
			}

			return cobra.ExactArgs(1)(cmd, args)
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			currDir, err := os.Getwd()
			if err != nil {
				return err
			}

			if abort {
				return fit_cmd.MergeAbort(currDir)
			}

			return fit_cmd.Merge(
				currDir,
				internal.CommitID(args[0]),
			)
		},
	}

	cmd.Flags().BoolVar(&abort, "abort", false, "Abort the current merge")

	return cmd
}

func NewCheckoutCmd() *cobra.Command {
	return &cobra.Command{
		Use: "checkout [commit]",

		Short: "Checkout a specific commit",

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
