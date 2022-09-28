package cmd

import (
	"time"

	"github.com/myfintech/ark/src/go/lib/ark/workspace"
	"github.com/myfintech/ark/src/go/lib/fs/observer"
	"github.com/myfintech/ark/src/go/lib/logz"

	"github.com/myfintech/ark/src/go/lib/fs"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newCheckIgnoreCmd(
	checkCmd *cobra.Command,
	config *workspace.Config,
	fileObserver *observer.Observer,
	logger *logz.Writer,
) *cobra.Command {
	var checkIgnoreCmd = &cobra.Command{
		Use:   "ignore",
		Short: "ark check ignore allows you to assert if a file will be ignored by ark",
		PreRunE: cobraRunEMiddleware(
			runE(cobra.MinimumNArgs(1)),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := fs.NormalizePath(config.Root(), args[0])
			if err != nil {
				return err
			}
			select {
			case result := <-fileObserver.FileSystemStream.Observe():
				for _, file := range result.V.([]*fs.File) {
					if file.Name == path {
						logger.Infof("%s was not ignored", path)
						return nil
					}
				}
				logger.Warnf("%s was ignored", path)
				return nil
			case <-time.After(time.Second * 5):
				return errors.New("failed to scan FS within 5 seconds")
			}
		},
	}
	checkCmd.AddCommand(checkIgnoreCmd)
	return checkIgnoreCmd
}
