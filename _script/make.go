package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	version string
)

func main() {
	ldflags := fmt.Sprintf("-s -w -X 'main.BuildTime=%s' -X 'main.ProgramCommitID=%s' -X 'main.ProgramVersion=%s' -buildid=",
		getBuildTime(),
		getCommitID(),
		getVersion(),
	)

	cmd := exec.Command("go",
		"build",
		"-trimpath",
		"-ldflags",
		ldflags,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", string(output))
		panic(err)
	}
}

func init() {
	flag.StringVar(&version, "version", "", "set as `ProgramVersion` while not empty")
	flag.Func("goos", "set as env:`GOOS` while not empty", setGOOS)
	flag.Func("goarm", "set as env:`GOARM` while not empty", setGOARM)
	flag.Func("goarch", "set as env:`GOARCH` while not empty", setGOARCH)

	flag.Parse()
}

func setGOOS(goos string) error {
	if goos == "" {
		return nil
	}
	return os.Setenv("GOOS", goos)
}

func setGOARM(goarm string) error {
	if goarm == "" {
		return nil
	}
	return os.Setenv("GOARM", goarm)
}

func setGOARCH(goarch string) error {
	if goarch == "" {
		return nil
	}
	return os.Setenv("GOARCH", goarch)
}

// getBuildTime get UTC time with "+%F-%Z/%T" format
func getBuildTime() string {
	return time.Now().UTC().Format("2006-01-02-UTC/15:04:05")
}

func getCommitID() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")

	if output, err := cmd.Output(); err != nil {
		return ""
	} else {
		return strings.TrimSpace(string(output)) // remove the tailing '\n'
	}
}

// getVersion return programVersion for ldflags
// using version while provided by os.Args
// or using latest tag provided by git
func getVersion() string {
	if version != "" {
		return version
	}

	cmd := exec.Command("git", "describe", "--tags")

	if output, err := cmd.Output(); err != nil {
		return ""
	} else {
		return strings.TrimSpace(string(output)) // remove the tailing '\n'
	}
}
