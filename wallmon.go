package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"runtime"
	"time"
)

const (
	minRestartInterval = 2 * time.Second
	minRestartDelay    = 1 * time.Second
)

type config struct {
	dataDir         string
	uri             string
	chromeCmd       string
	restartInterval time.Duration
	restartDelay    time.Duration
}

func main() {
	var (
		cnf config
		err error
	)
	if cnf, err = parseAndValidateFlags(); err != nil {
		fmt.Printf("invalid config: %s\n", err)
		os.Exit(1)
	}
	if err := initializeDataDir(cnf.dataDir); err != nil {
		fmt.Printf("unable to initialize data dir '%s': %s\n", cnf.dataDir, err)
		os.Exit(1)
	}
	dirArg := fmt.Sprintf("--user-data-dir=%s", cnf.dataDir)
	args := []string{"--kiosk", dirArg, cnf.uri}
	for {
		timeout, _ := context.WithTimeout(context.Background(), cnf.restartInterval)
		if err := runContext(timeout, cnf.chromeCmd, args...); err != nil {
			fmt.Printf("failed to run command: %s\n", err)
			os.Exit(1)
		}
		time.Sleep(cnf.restartDelay)
	}
}

func runContext(ctx context.Context, cmdName string, args ...string) error {
	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to start command: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) && exitErr.ExitCode() != -1 {
			return fmt.Errorf("command failed unexpectedly: %w", err)
		}
	}
	return nil
}

func initializeDataDir(dataDir string) error {
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		fmt.Printf("data dir '%s' does not exist, creating\n", dataDir)
		err = os.Mkdir(dataDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("creating dataDir: %w", err)
		}
	}
	f, err := os.OpenFile(path.Join(dataDir, "First Run"), os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("create First Run file: %w", err)
	}
	defer f.Close()
	return nil

}

func parseAndValidateFlags() (config, error) {
	var cnf config
	flag.StringVar(&cnf.dataDir, "data-dir", path.Join(os.TempDir(), "wallmon-data-dir"), "the data-directory to use for chrome")
	flag.StringVar(&cnf.uri, "url", "", "the uri to visit")
	flag.DurationVar(&cnf.restartInterval, "restart-interval", 12*time.Hour, fmt.Sprintf("restart interval of chrome (min %s)", minRestartInterval))
	flag.DurationVar(&cnf.restartDelay, "restart-delay", 1*time.Second, fmt.Sprintf("delay between restarts of chrome (min %s)", minRestartDelay))
	flag.StringVar(&cnf.chromeCmd, "chrome-cmd", defaultChromeCmdName(), "path to chrome cmd")
	flag.Parse()
	if cnf.chromeCmd == "" {
		return config{}, fmt.Errorf("chrome path needs to be specified")
	}
	if _, err := exec.LookPath(cnf.chromeCmd); err != nil {
		return config{}, fmt.Errorf("invalid chrome cmd: %w", err)
	}
	if cnf.dataDir == "" {
		return config{}, fmt.Errorf("data-dir can not be empty")
	}
	if cnf.uri == "" {
		return config{}, fmt.Errorf("url must be defined")
	}
	if _, err := url.ParseRequestURI(cnf.uri); err != nil {
		return config{}, fmt.Errorf("unable to validate url: %w", err)
	}
	if cnf.restartInterval < minRestartInterval {
		return config{}, fmt.Errorf("restart interval needs to be greater than: %s", minRestartInterval)
	}
	if cnf.restartDelay < minRestartDelay {
		return config{}, fmt.Errorf("restart delay needs to be greater than: %s", minRestartInterval)
	}
	return cnf, nil
}

func defaultChromeCmdName() string {
	switch runtime.GOOS {
	case "darwin":
		return firstExistingCmd("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome")
	case "windows":
		return firstExistingCmd("\\Program Files (x86)\\Google\\Chrome\\Application\\chrome")
	case "linux":
		return firstExistingCmd("google-chrome", "chromium-browser")
	}
	return ""
}

func firstExistingCmd(cmds ...string) string {
	for _, cmd := range cmds {
		if _, err := exec.LookPath(cmd); err == nil {
			return cmd
		}
	}
	return ""
}
