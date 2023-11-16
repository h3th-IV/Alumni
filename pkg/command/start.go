package command

import (
	"github.com/jim-nnamdi/bashfans/pkg/runner"
	"github.com/urfave/cli/v2"
)

func StartCommand() *cli.Command {
	var (
		startRunner = &runner.StartRunner{}
	)

	cmd := &cli.Command{
		Name:  "start",
		Usage: "starts the patron server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "listen-addr",
				EnvVars:     []string{"LISTEN_ADDR"},
				Usage:       "the address that the patron server will listen for request on",
				Destination: &startRunner.ListenAddr,
				Value:       ":8080", // TODO: check that this is correct port to serve on
			},
		},
		Action: startRunner.Run,
	}
	return cmd
}
