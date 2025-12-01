package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

// Version set at compile-time
var Version = "dev"

const asciiArt = `
________                                            ____.              __   .__
\______ \_______  ____   ____   ____               |    | ____   ____ |  | _|__| ____   ______
 |    |  \_  __ \/  _ \ /    \_/ __ \   ______     |    |/ __ \ /    \|  |/ /  |/    \ /  ___/
 |    |   \  | \(  <_> )   |  \  ___/  /_____/ /\__|    \  ___/|   |  \    <|  |   |  \\___ \
/_______  /__|   \____/|___|  /\___  >         \________|___  >___|  /__|_ \__|___|  /____  >
        \/                  \/     \/                        \/     \/     \/       \/     \/
                                                                    version: {{.Version}}
`

func main() {
	// Load env-file if it exists first
	if filename, found := os.LookupEnv("PLUGIN_ENV_FILE"); found {
		if err := godotenv.Load(filename); err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: failed to load env file %s: %v", filename, err)
		}
	}

	if _, err := os.Stat("/run/drone/env"); err == nil {
		if err := godotenv.Overload("/run/drone/env"); err != nil {
			log.Printf("Warning: failed to load /run/drone/env: %v", err)
		}
	}

	app := cli.NewApp()
	app.Name = "jenkins plugin"
	app.Usage = "trigger jenkins jobs"
	app.Copyright = "Copyright (c) 2019 Bo-Yi Wu"
	app.Authors = []*cli.Author{
		{
			Name:  "Bo-Yi Wu",
			Email: "appleboy.tw@gmail.com",
		},
	}
	app.Action = run
	app.Version = Version
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "host",
			Usage:   "jenkins base url",
			EnvVars: []string{"PLUGIN_URL", "JENKINS_URL", "INPUT_URL"},
		},
		&cli.StringFlag{
			Name:    "user",
			Aliases: []string{"u"},
			Usage:   "jenkins username",
			EnvVars: []string{"PLUGIN_USER", "JENKINS_USER", "INPUT_USER"},
		},
		&cli.StringFlag{
			Name:    "token",
			Aliases: []string{"t"},
			Usage:   "jenkins API token for authentication",
			EnvVars: []string{"PLUGIN_TOKEN", "JENKINS_TOKEN", "INPUT_TOKEN"},
		},
		&cli.StringFlag{
			Name:    "remote-token",
			Usage:   "jenkins remote trigger token",
			EnvVars: []string{"PLUGIN_REMOTE_TOKEN", "JENKINS_REMOTE_TOKEN", "INPUT_REMOTE_TOKEN"},
		},
		&cli.StringSliceFlag{
			Name:    "job",
			Aliases: []string{"j"},
			Usage:   "jenkins job",
			EnvVars: []string{"PLUGIN_JOB", "JENKINS_JOB", "INPUT_JOB"},
		},
		&cli.BoolFlag{
			Name:    "insecure",
			Usage:   "allow insecure server connections when using SSL",
			EnvVars: []string{"PLUGIN_INSECURE", "JENKINS_INSECURE", "INPUT_INSECURE"},
		},
		&cli.StringSliceFlag{
			Name:    "parameters",
			Aliases: []string{"p"},
			Usage:   "jenkins build parameters",
			EnvVars: []string{"PLUGIN_PARAMETERS", "JENKINS_PARAMETERS", "INPUT_PARAMETERS"},
		},
	}

	// Override a template
	cli.AppHelpTemplate = asciiArt + `
NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} ` +
		`{{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
COMMANDS:
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}` +
		`{{if .VisibleFlags}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}{{if .Copyright }}
COPYRIGHT:
   {{.Copyright}}
   {{end}}{{if .Version}}
VERSION:
   {{.Version}}
   {{end}}
REPOSITORY:
    Github: https://github.com/appleboy/drone-jenkins
`

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	// Validate required parameters
	if c.String("host") == "" {
		return fmt.Errorf("host is required")
	}

	if len(c.StringSlice("job")) == 0 {
		return fmt.Errorf("at least one job is required")
	}

	// Validate authentication: either (user + token) or remote-token must be provided
	hasUserAuth := c.String("user") != "" && c.String("token") != ""
	hasRemoteToken := c.String("remote-token") != ""

	if !hasUserAuth && !hasRemoteToken {
		return fmt.Errorf("authentication required: provide either (user + token) or remote-token")
	}

	plugin := Plugin{
		BaseURL:     c.String("host"),
		Username:    c.String("user"),
		Token:       c.String("token"),
		RemoteToken: c.String("remote-token"),
		Job:         c.StringSlice("job"),
		Insecure:    c.Bool("insecure"),
		Parameters:  c.StringSlice("parameters"),
	}

	return plugin.Exec()
}
