package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/urfave/cli/v2"
)

const (
	version string = "0.1.0"
)

// Run
func main() {
	// Initialise application
	app := cli.NewApp()

	// Set application info
	app.Name = "Web3.Storage CLI Relayer"
	app.Usage = "Relays a specific subset of commands to the Web3.Storage CLI via a Unix Socket"
	app.Version = version
	app.Authors = []*cli.Author{
		{
			Name:  "Joe Clapis",
			Email: "joe@rocketpool.net",
		},
	}
	app.Copyright = "(C) 2023 Rocket Pool Pty Ltd"

	app.ArgsUsage = "some args go here"

	// Set application flags
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "watchtower-socket-path",
			Aliases: []string{"w"},
			Usage:   "The path of the Unix socket to host the Watchtower's relayer on (required)",
		},
		&cli.StringFlag{
			Name:    "cli-socket-path",
			Aliases: []string{"c"},
			Usage:   "The path of the Unix socket to host the CLI's relayer on (optional)",
		},
		&cli.StringFlag{
			Name:    "rewards-files-path",
			Aliases: []string{"r"},
			Usage:   "The path of the mounted rewards files volume (required)",
		},
	}

	app.Action = func(c *cli.Context) error {
		watchtowerSocketPath := c.String("watchtower-socket-path")
		if watchtowerSocketPath == "" {
			return fmt.Errorf("socket path is required to run")
		}
		rewardsFilesPath := c.String("rewards-files-path")
		if rewardsFilesPath == "" {
			return fmt.Errorf("rewards files path is required to run")
		}
		cliSocketPath := c.String("cli-socket-path")

		// Wait group to handle the relayers
		wg := new(sync.WaitGroup)
		wg.Add(1)

		// Start the CLI relayer
		var cliRelayer *relayManager
		if cliSocketPath != "" {
			if cliSocketPath == watchtowerSocketPath {
				return fmt.Errorf("cli socket path cannot be the same as the watchtower socket path")
			}
			cliRelayer = newRelayManager(cliSocketPath, rewardsFilesPath)
			wg.Add(1)
			err := cliRelayer.start(wg)
			if err != nil {
				return fmt.Errorf("error starting CLI relayer: %w", err)
			}
		}

		// Start the watchtower relayer
		watchtowerRelayer := newRelayManager(watchtowerSocketPath, rewardsFilesPath)
		err := watchtowerRelayer.start(wg)
		if err != nil {
			return fmt.Errorf("error starting watchtower relayer: %w", err)
		}

		// Handle process closures
		termListener := make(chan os.Signal, 1)
		signal.Notify(termListener, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-termListener
			watchtowerRelayer.stop()
			if cliRelayer != nil {
				cliRelayer.stop()
			}
			os.Exit(0)
		}()

		// Run the relayers until closed
		wg.Wait()
		return nil
	}

	// Run application
	fmt.Println("")
	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Error running relayer: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Println("")
}
