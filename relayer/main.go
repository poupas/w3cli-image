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
			Name:    "socket-path",
			Aliases: []string{"s"},
			Usage:   "The path of the Unix socket to host the Watchtower's relayer on (required)",
		},
		&cli.StringFlag{
			Name:    "rewards-files-path",
			Aliases: []string{"r"},
			Usage:   "The path of the mounted rewards files volume (required)",
		},
	}

	app.Action = func(c *cli.Context) error {
		socketPath := c.String("socket-path")
		if socketPath == "" {
			return fmt.Errorf("socket path is required to run")
		}
		rewardsFilesPath := c.String("rewards-files-path")
		if rewardsFilesPath == "" {
			return fmt.Errorf("rewards files path is required to run")
		}

		// Wait group to handle the relayers
		wg := new(sync.WaitGroup)
		wg.Add(1)

		// Start the watchtower relayer
		watchtowerRelayer := newRelayManager(socketPath, rewardsFilesPath)
		err := watchtowerRelayer.start(wg)
		if err != nil {
			return fmt.Errorf("error starting watchtower relayer: %w", err)
		}

		// Handle process closures
		termListener := make(chan os.Signal, 1)
		signal.Notify(termListener, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-termListener
			fmt.Println("Shutting down relayer...")
			err := watchtowerRelayer.stop()
			if err != nil {
				fmt.Printf("WARNING: relayer didn't shutdown cleanly: %s\n", err.Error())
				wg.Done()
			}
		}()

		// Run the relayers until closed
		fmt.Printf("Started relayer on %s.\n", watchtowerRelayer.socketPath)
		wg.Wait()
		fmt.Println("Relayer stopped.")
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
