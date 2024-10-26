package app

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/Alhanaqtah/caching-proxy/internal/service"

	"github.com/urfave/cli/v2"
)

func Run(ctx context.Context, args []string) error {
	app := cli.NewApp()
	app.Name = "caching-proxy"
	app.Usage = "A proxy server for caching"
	app.Description = "Caching server that caches responses from other servers."
	app.UsageText = "caching-proxy [global options] command [command options]"

	app.Commands = []*cli.Command{
		{
			Name:        "run",
			Usage:       "Runs the server",
			Description: "Starts a proxy server for caching",
			Action: func(c *cli.Context) error {
				port := c.String("port")
				origin := c.String("origin")
				ttlStr := c.String("ttl")

				if _, err := strconv.Atoi(port); err != nil {
					log.Fatalf("invalid value for flag port: %v", err)
				}

				if _, err := url.ParseRequestURI(origin); err != nil {
					return fmt.Errorf("invalid origin URL: %v", err)
				}

				ttl, err := strconv.Atoi(ttlStr)
				if err != nil {
					return fmt.Errorf("invalid value for flag ttl: %v", err)
				}
				ttlDuration := time.Duration(ttl) * time.Second

				return service.Run(port, origin, ttlDuration)
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "port",
					Aliases:  []string{"p"},
					Required: true,
					Usage:    "Port on which the server will listen",
				},
				&cli.StringFlag{
					Name:     "origin",
					Aliases:  []string{"o"},
					Required: true,
					Usage:    "Origin server to cache responses from",
				},
				&cli.StringFlag{
					Name:     "ttl",
					Aliases:  []string{"t"},
					Required: true,
					Usage:    "Cache items ttl in seconds",
				},
			},
		},
	}

	return app.RunContext(ctx, args)
}
