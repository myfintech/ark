package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	prompt "github.com/AlecAivazis/survey/v2"

	"github.com/myfintech/ark/src/go/lib/log"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/tools/arkcli/cmd/constants/workspacebuildfile"
)

const (
	promptErrorMessage         = "unable to successfully ask question and/or record answer"
	buildHCLWriterErrorMessage = "unable to write to BUILD.hcl file"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes workspace building with ark",
	Long: `Initializes a workspace for building with ark.

It first checks to see if the workspace is already initialized,
prompts the user for some initial input, and creates a set of initial
files used by ark for intelligent building.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		workspaceCheck := base.NewWorkspace()
		if err := workspaceCheck.DetermineRootFromCWD(); err == nil {
			log.Infoln("the workspace has already been initialized")
			return nil
		}

		currentDir, err := os.Getwd()
		if err != nil {
			return errors.Wrap(err, "unable to get current directory")
		}

		// placeholders for the answers to our questions
		var (
			runSetup                bool
			packageManagerSelection []string
		)
		// the questions to adk the user
		setupWorkspace := &prompt.Confirm{
			Message: fmt.Sprintf("Setup workspace at %s", currentDir),
			Default: false,
		}
		packageManager := &prompt.MultiSelect{
			Message: "What package manager does your service use?",
			Options: []string{"JavaScript - lerna", "Golang - mod"},
		}

		// boilerFiles are the bare minimum files that need to be present when one starts to use the ark cli
		var boilerFiles = []string{
			"WORKSPACE.hcl",
			"BUILD.hcl",
		}

		if err = prompt.AskOne(setupWorkspace, &runSetup); err != nil {
			return errors.Wrap(err, promptErrorMessage)
		}
		if !runSetup {
			return nil
		}

		if err = prompt.AskOne(packageManager, &packageManagerSelection); err != nil {
			return errors.Wrap(err, promptErrorMessage)
		}
		log.Debugf("option(s) chosen: %s", strings.Join(packageManagerSelection, ", "))

		for _, file := range boilerFiles {
			if _, err = os.Stat(file); os.IsNotExist(err) {
				log.Infof("%s does not exist; creating ...", file)
				f, createErr := os.Create(file)
				if createErr != nil {
					return errors.Wrap(err, fmt.Sprintf("unable to create file: %s", file))
				}
				if err = f.Close(); err != nil {
					return errors.Wrap(err, fmt.Sprintf("unalbe to close new file: %s", file))
				}
			}
		}

		// write the BUILD.hcl file for the workspace
		workspaceBuildFileRef, err := os.OpenFile("BUILD.hcl", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return errors.Wrap(err, "unable to open BUILD.hcl file")
		}
		defer workspaceBuildFileRef.Close()
		if _, err = workspaceBuildFileRef.WriteString(workspacebuildfile.Package); err != nil {
			return errors.Wrap(err, buildHCLWriterErrorMessage)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
