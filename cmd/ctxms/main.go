package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rotationalio/ctxms"
	api "github.com/rotationalio/ctxms/proto"
	cli "github.com/urfave/cli/v2"
	"google.golang.org/grpc"
)

func main() {
	app := cli.NewApp()
	app.Name = "ctxms"
	app.Usage = "context microservice management test code"
	app.Version = "beta"
	app.Flags = []cli.Flag{}

	app.Commands = []*cli.Command{
		{
			Name:     "serve",
			Usage:    "run the ctmx server",
			Category: "server",
			Action:   serve,
			Flags: []cli.Flag{
				&cli.UintFlag{
					Name:    "port",
					Aliases: []string{"p"},
					Usage:   "port to listen for requests on; forwards to next port",
					Value:   9000,
				},
				&cli.DurationFlag{
					Name:    "delay",
					Aliases: []string{"d"},
					Usage:   "maximum delay before forwarding",
					Value:   10 * time.Second,
				},
				&cli.BoolFlag{
					Name:    "terminal",
					Aliases: []string{"t"},
					Usage:   "is the terminal server in the hop chain",
				},
				&cli.StringFlag{
					Name:    "name",
					Aliases: []string{"n"},
					Usage:   "unique name of the server",
				},
			},
		},
		{
			Name:     "trace",
			Usage:    "run an multi-hop trace route",
			Category: "client",
			Action:   trace,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "endpoint",
					Aliases: []string{"e"},
					Usage:   "endpoint to connect to the server on",
					Value:   "localhost:9000",
				},
				&cli.DurationFlag{
					Name:    "timeout",
					Aliases: []string{"t"},
					Usage:   "set deadline for client-side context",
					Value:   30 * time.Second,
				},
			},
		},
	}

	app.Run(os.Args)
}

func serve(c *cli.Context) (err error) {
	conf := ctxms.NewConfig()
	conf.Name = c.String("name")
	conf.Port = uint16(c.Uint("port"))
	conf.Delay = c.Duration("delay")
	conf.Terminal = c.Bool("terminal")

	var server *ctxms.Server
	if server, err = ctxms.New(conf); err != nil {
		return cli.Exit(err, 1)
	}

	if err = server.Serve(); err != nil {
		return cli.Exit(err, 1)
	}
	return nil
}

func trace(c *cli.Context) (err error) {
	var cc *grpc.ClientConn
	if cc, err = grpc.Dial(c.String("endpoint"), grpc.WithInsecure()); err != nil {
		return cli.Exit(err, 1)
	}
	defer cc.Close()

	client := api.NewHopperClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), c.Duration("timeout"))
	defer cancel()

	pkt := &api.Packet{
		Id:        uuid.NewString(),
		Timestamp: time.Now().Format(time.RFC3339Nano),
	}

	if pkt, err = client.Trace(ctx, pkt); err != nil {
		return cli.Exit(err, 1)
	}

	fmt.Println(pkt.Repr())
	return nil
}
