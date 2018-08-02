package main

import (
	"os"

	"github.com/bbengfort/livenet"
	"github.com/joho/godotenv"
	"github.com/urfave/cli"
)

func main() {

	// Load the .env file if it exists
	godotenv.Load()

	// Instantiate the command line application
	app := cli.NewApp()
	app.Name = "livenet"
	app.Version = livenet.PackageVersion
	app.Usage = "run the livenet service"

	// Define commands available to application
	app.Commands = []cli.Command{
		{
			Name:     "serve",
			Usage:    "run a livenet server",
			Action:   serve,
			Category: "server",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "n, name",
					Usage: "specify the unique hostname for the server",
					Value: "",
				},
				cli.StringFlag{
					Name:  "c, config",
					Usage: "configuration file for server",
					Value: "config.json",
				},
				cli.DurationFlag{
					Name:  "u, uptime",
					Usage: "specify a duration for the server to run",
					Value: 0,
				},
			},
		},
		// {
		// 	Name:     "commit",
		// 	Usage:    "commit an entry to the distributed log",
		// 	Before:   initConfig,
		// 	Action:   commit,
		// 	Category: "client",
		// 	Flags: []cli.Flag{
		// 		cli.StringFlag{
		// 			Name:  "c, config",
		// 			Usage: "configuration file for network",
		// 			Value: "",
		// 		},
		// 		cli.StringFlag{
		// 			Name:  "k, key",
		// 			Usage: "the name of the command to commit",
		// 		},
		// 		cli.StringFlag{
		// 			Name:  "v, value",
		// 			Usage: "the value of the command to commit",
		// 		},
		// 	},
		// },
		// {
		// 	Name:     "bench",
		// 	Usage:    "run a raft benchmark with concurrent network",
		// 	Before:   initConfig,
		// 	Action:   bench,
		// 	Category: "client",
		// 	Flags: []cli.Flag{
		// 		cli.StringFlag{
		// 			Name:  "c, config",
		// 			Usage: "configuration file for replica",
		// 			Value: "",
		// 		},
		// 		cli.IntFlag{
		// 			Name:  "n, nclients",
		// 			Usage: "number of concurrent clients to run",
		// 			Value: 4,
		// 		},
		// 		cli.Uint64Flag{
		// 			Name:  "r, requests",
		// 			Usage: "number of requests issued per client",
		// 			Value: 1000,
		// 		},
		// 	},
		// },
	}

	// Run the CLI program
	app.Run(os.Args)
}

//===========================================================================
// Initialization
//===========================================================================

//===========================================================================
// Server Commands
//===========================================================================

func serve(c *cli.Context) (err error) {

	conf := new(livenet.Config)
	if err = conf.Load(c.String("config")); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	conf.Name = c.String("name")

	server, err := livenet.New(conf)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	if err = server.Listen(); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	// if err = replica.Listen(); err != nil {
	// 	return cli.NewExitError(err.Error(), 1)
	// }

	return nil
}

//===========================================================================
// Client Commands
//===========================================================================

// func commit(c *cli.Context) (err error) {
// 	if client, err = raft.NewClient(config); err != nil {
// 		return cli.NewExitError(err.Error(), 1)
// 	}
//
// 	var entry *pb.LogEntry
// 	if entry, err = client.Commit(c.String("key"), []byte(c.String("value"))); err != nil {
// 		return cli.NewExitError(err.Error(), 1)
// 	}
//
// 	fmt.Println(entry)
//
// 	return nil
// }
//
// func bench(c *cli.Context) error {
// 	benchmark, err := raft.NewBenchmark(
// 		config, c.Int("nclients"), c.Uint64("requests"),
// 	)
//
// 	if err != nil {
// 		return cli.NewExitError(err.Error(), 1)
// 	}
//
// 	fmt.Println(benchmark)
// 	return nil
// }
