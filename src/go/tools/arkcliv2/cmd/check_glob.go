package cmd

import (
	"fmt"

	gitignorev5 "github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/myfintech/ark/src/go/lib/ark/workspace"
	"github.com/myfintech/ark/src/go/lib/logz"

	"github.com/myfintech/ark/src/go/lib/git/gitignore"

	"github.com/myfintech/ark/src/go/lib/fs"
	"github.com/spf13/cobra"
)

func newCheckGlobCmd(
	checkCmd *cobra.Command,
	logger *logz.Writer,
	config *workspace.Config,
	ignorePatterns []gitignorev5.Pattern,
) *cobra.Command {
	var checkGlobCmd = &cobra.Command{
		Use:   "pattern",
		Short: "ark check pattern allows you to query for files that match a given pattern",
		PreRunE: cobraRunEMiddleware(
			runE(cobra.MinimumNArgs(1)),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			pattern, err := fs.NormalizePath(config.Root(), args[0])
			if err != nil {
				return err
			}
			logger.Debugf("pattern %s", args[0])
			logger.Debugf("expended to %s", pattern)
			matches, err := fs.Glob(
				pattern,
				config.Root(),
				gitignore.NewMatcher(ignorePatterns),
			)
			if err != nil {
				return err
			}
			for _, match := range matches {
				fmt.Println(match)
			}
			return nil
		},
	}

	checkCmd.AddCommand(checkGlobCmd)
	return checkGlobCmd
}
