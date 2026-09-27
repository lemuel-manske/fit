package fit

import (
	"fmt"
	"os"
	"strings"

	"path/filepath"

	"github.com/spf13/cobra"
)

func InitCLI(root *cobra.Command) {
	root.AddCommand(NewAddCmd())
	root.AddCommand(NewCommitCmd())
	root.AddCommand(NewInitCmd())
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
		Use:   "fit",
	}
}

func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",

		Args: cobra.ExactArgs(0),

		RunE: func(cmd *cobra.Command, args []string) error {
			currDir, err := GetCurrentDirectory()
			if err != nil {
				return err
			}

			repositoryName := strings.TrimSuffix(filepath.Base(currDir), filepath.Ext(currDir))

			return Init(currDir, repositoryName)
		},
	}
}

func NewAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add",

		Args: cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			currDir, err := GetCurrentDirectory()
			if err != nil {
				return err
			}

			return Add(currDir, args[0])
		},
	}
}

func NewRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm",

		Args: cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			currDir, err := GetCurrentDirectory()
			if err != nil {
				return err
			}

			return Rm(currDir, args[0])
		},
	}
}

func NewCommitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "commit",

		Args: cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			currDir, err := GetCurrentDirectory()
			if err != nil {
				return err
			}

			commitID, err := CommitChanges(currDir, args[0])

			if err != nil {
				return err
			}

			fmt.Println(commitID)
			return nil
		},
	}
}
