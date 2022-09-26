package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/myfintech/ark/src/go/lib/fs"
	"github.com/spf13/viper"

	"github.com/myfintech/ark/src/go/lib/log"
	"github.com/myfintech/ark/src/go/lib/pkg"
	"github.com/myfintech/ark/src/go/tools/arkcli/cmd"
)

func main() {
	cleanup := checkAndStartCPUProfiling()
	defer cleanup()

	checkAndStartMemoryProfiling()

	cmd.Execute()
	for _, i := range os.Args {
		if i == "__complete" || i == "upgrade" {
			return
		}
	}
	err := pkg.VersionCheckHook(viper.GetBool("skip_version_check"), func(current, latest pkg.PackageInfo) error {
		fmt.Printf("Your version of arkcli is: %s.\nThe latest version is: %s.\nRun `ark upgrade` to install the latest version.\n", current.Version, latest.Version)
		return nil
	}, nil)
	if err != nil {
		log.Warn(err)
	}
}

func checkAndStartCPUProfiling() func() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	_ = cmd.RootCmd().PersistentFlags().Parse(os.Args[1:])
	cpuProfile, err := cmd.RootCmd().PersistentFlags().GetString("cpu_profile")
	if err != nil {
		panic(err)
	}
	if cpuProfile == "" {
		return func() {}
	}

	cpuProfileOutPath, err := fs.NormalizePath(cwd, cpuProfile)
	if err != nil {
		panic(err)
	}
	cpuProfileOut, err := os.Create(cpuProfileOutPath)
	if err != nil {
		panic(err)
	}
	if err = pprof.StartCPUProfile(cpuProfileOut); err != nil {
		panic(err)
	}
	fmt.Printf("writing cpu profile to %s\n", cpuProfileOutPath)
	return pprof.StopCPUProfile
}

func checkAndStartMemoryProfiling() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	_ = cmd.RootCmd().PersistentFlags().Parse(os.Args[1:])
	memProfile, err := cmd.RootCmd().PersistentFlags().GetString("mem_profile")
	if err != nil {
		panic(err)
	}
	if memProfile == "" {
		return
	}

	memProfileOutPath, err := fs.NormalizePath(cwd, memProfile)
	if err != nil {
		panic(err)
	}
	memProfileOut, err := os.Create(memProfileOutPath)
	if err != nil {
		panic(err)
	}

	runtime.GC()
	if err = pprof.WriteHeapProfile(memProfileOut); err != nil {
		panic(err)
	}
	fmt.Printf("writing memory profile to %s\n", memProfileOutPath)
}
