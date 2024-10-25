package app

import (
	"context"

	"github.com/urfave/cli/v2"
)

func Run(ctx context.Context, args []string) error {
	app := cli.NewApp()

	app.Name = "caching-proxy"
	app.Usage = "A proxy server for caching"
	app.Description = "caching server that caches responses from other servers."

	app.Commands = []*cli.Command{
		{
			Name:        "run",
			Usage:       "Runs the server",
			Description: "starts a proxy server for caching",
			Action: func(ctx *cli.Context) error {
				return nil
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "port",
					Aliases:  []string{"p"},
					Value:    "8080",
					Required: true,
					Usage:    "port on which the server will listen",
				},
				&cli.StringFlag{
					Name:     "origin",
					Aliases:  []string{"o"},
					Required: true,
					Usage:    "origin server to cache responses from",
				},
			},
		},
		{
			Name:        "clear",
			Aliases:     []string{"c"},
			Usage:       "Removes all cached responses",
			Description: "removes all cached responses",
			Action: func(ctx *cli.Context) error {
				return nil
			},
		},
	}

	return app.RunContext(ctx, args)
}
