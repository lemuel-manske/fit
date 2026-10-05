package fit

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	fit_cmd "fit/fit/cmd"

	"fit/fit/internal"

	"path/filepath"

	"github.com/spf13/cobra"
)

func InitCLI(root *cobra.Command) {
	root.AddCommand(NewAddCmd())
	root.AddCommand(NewCheckoutCmd())
	root.AddCommand(NewCloneCmd())
	root.AddCommand(NewCommitCmd())
	root.AddCommand(NewInitCmd())
	root.AddCommand(NewMergeCmd())
	root.AddCommand(NewReposCmd())
	root.AddCommand(NewRmCmd())
	root.AddCommand(NewServeCmd())
	root.AddCommand(NewStatusCmd())
	root.AddCommand(NewSyncCmd())
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

			return fit_cmd.InitNew(currDir, repositoryName)
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

	cmd.Flags().StringVarP(
		&message,
		"message",
		"m",
		"",
		"Commit message",
	)

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

	cmd.Flags().BoolVar(
		&abort,
		"abort",
		false,
		"Abort the current merge",
	)

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

func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use: "status",

		Short: "Show the status of the repository",

		Args: cobra.ExactArgs(0),

		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return err
			}

			status, err := fit_cmd.GetStatus(dir)
			if err != nil {
				return err
			}

			status.Print(cmd.OutOrStdout())

			return nil
		},
	}
}

func NewServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Serve this FIT peer",
		Args:  cobra.NoArgs,

		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return err
			}

			return fit_cmd.Serve(
				dir,
				cmd.Context(),
			)
		},
	}
}

func NewReposCmd() *cobra.Command {
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "repos",
		Short: "Discover FIT repositories",
		Args:  cobra.NoArgs,

		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return err
			}

			config, err := internal.LoadConfig(dir)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(
				cmd.Context(),
				timeout,
			)
			defer cancel()

			offers, err := fit_cmd.DiscoverAll(dir, ctx)
			if err != nil {
				return err
			}

			seen := map[internal.RepositoryID]bool{}

			for offer := range offers {
				if seen[offer.RepositoryID] {
					continue
				}

				seen[offer.RepositoryID] = true

				out := cmd.OutOrStdout()

				name := offer.RepositoryName

				if offer.PeerID == config.PeerID {
					name += " (You)"
				}

				fmt.Fprintf(out, "  %s\n", name)
				fmt.Fprintf(out, "    repository: %s\n", offer.RepositoryID)
				fmt.Fprintf(out, "    peer:       %s\n", offer.PeerID)

				if offer.Head != "" {
					fmt.Fprintf(out, "    head:       %s\n", offer.Head)
				}

				fmt.Fprintln(out)
			}

			return nil
		},
	}

	cmd.Flags().DurationVar(
		&timeout,
		"timeout",
		2*time.Second,
		"Discovery timeout",
	)

	return cmd
}

func NewSyncCmd() *cobra.Command {
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize objects from remote peers",
		Args:  cobra.NoArgs,

		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(
				cmd.Context(),
				timeout,
			)
			defer cancel()

			return fit_cmd.Sync(dir, ctx)
		},
	}

	cmd.Flags().DurationVar(
		&timeout,
		"timeout",
		10*time.Second,
		"Sync timeout",
	)

	return cmd
}

func NewCloneCmd() *cobra.Command {
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "clone <repository> <directory>",
		Short: "Clone a FIT repository",
		Args:  cobra.ExactArgs(2),

		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(
				cmd.Context(),
				timeout,
			)
			defer cancel()

			return fit_cmd.Clone(
				ctx,
				args[0],
				args[1],
			)
		},
	}

	cmd.Flags().DurationVar(
		&timeout,
		"timeout",
		30*time.Second,
		"Clone timeout",
	)

	return cmd
}
