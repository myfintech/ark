package cmd

import (
	"bytes"
	"fmt"
	"os"
	osexec "os/exec"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/myfintech/ark/src/go/lib/exec"
)

const debugGraphDescription = `
graph is a sub-command of debug that outputs the ark graph or the sub graph of a specific target

To use the --format png flag you must install graphviz which comes with the dot command

https://graphviz.org/download/

MacOS:
	$ brew install graphviz

Ubuntu / Debian packages
	$ sudo apt install graphviz

Fedora project*
	$ sudo yum install graphviz

Redhat Enterprise, or CentOS systems* available but are out of date.
	$ sudo yum install graphviz
`

var debugGraphCmd = &cobra.Command{
	Use:               "graph [TARGET_ADDRESS]",
	Short:             "a command to dump debugging information on the graph",
	Long:              debugGraphDescription,
	ValidArgsFunction: getValidTargets,
	RunE: func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("format")
		graph := workspace.TargetGraph

		if len(args) > 0 {
			vertex, err := workspace.TargetLUT.LookupByAddress(args[0])
			if err != nil {
				return err
			}
			graph = *graph.Isolate(vertex)
		}

		switch format {
		case "json":
			data, err := graph.MarshalJSON()
			if err != nil {
				return err
			}
			fmt.Print(string(data))
			break
		case "text":
			fmt.Println(graph.String())
			break
		case "dot":
			fmt.Println(string(graph.Dot()))
			break
		case "png":
			if len(workspace.TargetLUT) <= 1 {
				return errors.New("there must be at least two objects in the graph to render an image with graphviz")
			}
			dotCMD, err := osexec.LookPath("dot")
			if err != nil {
				return err
			}
			return exec.LocalExecutor(exec.LocalExecOptions{
				Command:          []string{dotCMD, "-Tpng", "-Gdpi=300"},
				Stdin:            bytes.NewBuffer(graph.Dot()),
				Stdout:           os.Stdout,
				Stderr:           os.Stderr,
				InheritParentEnv: true,
			}).Run()
		default:
			return errors.Errorf("%s is not a valid export format \n", format)
		}
		return nil
	},
}

func init() {
	debugCmd.AddCommand(debugGraphCmd)
	debugGraphCmd.Flags().StringP("format", "f", "dot", "The export format of the graph. Can be one of (dot|png|text|json)")
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
}
