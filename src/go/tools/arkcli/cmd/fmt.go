package cmd

import (
	"github.com/myfintech/ark/src/go/lib/hclutils"
	"github.com/myfintech/ark/src/go/lib/pattern"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "fmt is a sub-command of ark that formats HCL files",
	RunE: func(cmd *cobra.Command, args []string) error {
		list, err := cmd.Flags().GetBool("list")
		if err != nil {
			return err
		}
		write, err := cmd.Flags().GetBool("write")
		if err != nil {
			return err
		}
		diff, err := cmd.Flags().GetBool("diff")
		if err != nil {
			return err
		}
		check, err := cmd.Flags().GetBool("check")
		if err != nil {
			return err
		}
		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		if len(args) == 0 {
			args = append(args, ".")
		}

		matcher := &pattern.Matcher{
			Includes: []string{"**/BUILD.hcl", "**/WORKSPACE.hcl"},
			Excludes: []string{"**/vendor/**", "**/node_modules/**", "**/test*/**"},
		}
		if err = matcher.Compile(); err != nil {
			return err
		}

		opts := hclutils.FormatOpts{
			List:      list,
			Write:     write,
			Diff:      diff,
			Check:     check,
			Recursive: recursive,
			Input:     nil,
		}

		if diags := opts.Run(args[0], matcher); diags != nil && diags.HasErrors() {
			return diags
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(fmtCmd)
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
	fmtCmd.Flags().Bool("list", true, "List files whose formatting differs (always disabled if using STDIN)")
	fmtCmd.Flags().Bool("write", true, "Write to source files (always disabled if using STDIN)")
	fmtCmd.Flags().Bool("diff", false, "Display diffs of formatting changes")
	fmtCmd.Flags().Bool("check", false, "Check if the input is formatted. Exit status will be 0 if all input is properly formatted")
	fmtCmd.Flags().Bool("recursive", false, "Also process files in subdirectories")
}
