package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

func whoami() (string, bool, error) {
	result, err := runW3CliCommand("whoami")
	return result, false, err
}

func login(args url.Values) (string, bool, bool, error) {
	if !args.Has("email") {
		return "", false, true, fmt.Errorf("email parameter is required")
	}
	email := args.Get("email")
	result, err := runW3CliCommand("login", email)
	return result, false, false, err
}

func spaceCreate() (string, bool, error) {
	result, err := runW3CliCommand("space", "create", "rp_odao", "--no-recovery")
	return result, false, err
}

func up(args url.Values, rewardsFilesPath string) (string, bool, bool, error) {
	if !args.Has("file") {
		return "", false, true, fmt.Errorf("file parameter is required")
	}
	file := args.Get("file")

	// Make sure the file exists
	path := filepath.Join(rewardsFilesPath, file)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return "", false, true, fmt.Errorf("file '%s' does not exist", path)
	}

	result, err := runW3CliCommand("up", path, "--no-wrap", "--json")
	return result, true, false, err
}

// Runs a web3.storage CLI command
func runW3CliCommand(args ...string) (string, error) {
	cmd := exec.Command("w3", args...)
	stdout, err := cmd.Output()
	if err != nil {
		exitError, isExitError := err.(*exec.ExitError)
		if isExitError {
			return "", fmt.Errorf("w3 exited with status code %d:\n%s", exitError.ExitCode(), string(exitError.Stderr))
		}
		return "", err
	}
	return string(stdout), nil
}
