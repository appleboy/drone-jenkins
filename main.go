package main

import (
	"os"

	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/urfave/cli"
)

// Version set at compile-time
var Version string

func main() {
	app := cli.NewApp()
	app.Name = "jenkins plugin"
	app.Usage = "jenkins plugin"
	app.Action = run
	app.Version = Version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "base.url",
			Usage:  "jenkins base url",
			EnvVar: "PLUGIN_BASE_URL,JENKINS_BASE_URL",
		},
		cli.StringFlag{
			Name:   "username",
			Usage:  "jenkins username",
			EnvVar: "PLUGIN_USERNAME,JENKINS_USERNAME",
		},
		cli.StringFlag{
			Name:   "token",
			Usage:  "jenkins token",
			EnvVar: "PLUGIN_TOKEN,JENKINS_TOKEN",
		},
		cli.StringSliceFlag{
			Name:   "job",
			Usage:  "jenkins job",
			EnvVar: "PLUGIN_JOB,JENKINS_JOB",
		},
		cli.StringFlag{
			Name:   "env-file",
			Usage:  "source env file",
			EnvVar: "ENV_FILE",
		},
	}
	app.Run(os.Args)
}

func run(c *cli.Context) error {
	if c.String("env-file") != "" {
		_ = godotenv.Load(c.String("env-file"))
	}

	plugin := Plugin{
		BaseURL:  c.String("base.url"),
		Username: c.String("username"),
		Token:    c.String("token"),
		Job:      c.StringSlice("job"),
	}

	return plugin.Exec()
}
