package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript"

	"github.com/myfintech/ark/src/go/lib/logz"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

func _(
	rootCmd *cobra.Command,
	logger logz.FieldLogger,
	vm *typescript.VirtualMachine,
) *cobra.Command {
	var replCmd = &cobra.Command{
		Use:     "repl",
		Short:   "repl",
		Example: "ark repl",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Warn("ark repl is in alpha and not yet supported")
			l, err := readline.NewEx(&readline.Config{
				Prompt:      "\033[31m»\033[0m ",
				HistoryFile: "/tmp/readline.tmp",
				// AutoComplete:    completer,
				InterruptPrompt: "^C",
				EOFPrompt:       "exit",

				HistorySearchFold: true,
				// FuncFilterInputRune: filterInput,
			})
			if err != nil {
				panic(err)
			}

			defer func() {
				_ = l.Close()
			}()

			for {
				line, readE := l.Readline()
				if readE == readline.ErrInterrupt {
					if len(line) == 0 {
						break
					} else {
						continue
					}
				} else if readE == io.EOF {
					break
				}

				line = strings.TrimSpace(line)
				switch {
				case strings.HasPrefix(line, "import "):
					name := strings.TrimPrefix(line, "import ")
					logger.Debug("loading module", name)
					mod, err := vm.GetModule(name)
					if err != nil {
						fmt.Println(err)
						continue
					}
					if err = vm.SetRuntimeValue("_", mod); err != nil {
						fmt.Println(err)
						continue
					}
					continue
				}

				val, err := vm.RunScript("repl", line)
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Println(val.String())
			}
			return nil
		},
	}

	rootCmd.AddCommand(replCmd)
	return replCmd
}
