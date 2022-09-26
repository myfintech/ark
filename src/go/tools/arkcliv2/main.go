package main

import (
	_ "embed"
	"log"
	_ "net/http/pprof"
	"os"

	"github.com/myfintech/ark/src/go/lib/autoupdate"
	"github.com/myfintech/ark/src/go/tools/arkcliv2/cmd"
	"github.com/pkg/profile"
)

//go:embed version.json
var versionFile string

func main() {
	cli, err := cmd.BuildCLI()
	if err != nil {
		log.Fatalln(err)
	}

	mode, err := cli.RootCmd.PersistentFlags().GetString("profile-mode")
	if err != nil {
		log.Fatalln(err)
	}

	switch mode {
	case "":
		break
	case "cpu":
		defer profile.Start(profile.CPUProfile).Stop()
	case "mem":
		defer profile.Start(profile.MemProfile).Stop()
	case "mutex":
		defer profile.Start(profile.MutexProfile).Stop()
	case "block":
		defer profile.Start(profile.BlockProfile).Stop()
	default:
		log.Fatalf("invalid profile-mode %s", mode)
	}

	if err = autoupdate.InitFromString(versionFile); err != nil {
		log.Fatalln(err)
	}

	os.Exit(cli.Execute())
}
