package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/urfave/cli"
)

// Version set at compile-time
var Version string

func main() {
	// Load env-file if it exists first
	if filename, found := os.LookupEnv("PLUGIN_ENV_FILE"); found {
		_ = godotenv.Load(filename)
	}

	if _, err := os.Stat("/run/drone/env"); err == nil {
		_ = godotenv.Overload("/run/drone/env")
	}

	app := cli.NewApp()
	app.Name = "jenkins plugin"
	app.Usage = "trigger jenkins jobs"
	app.Copyright = "Copyright (c) 2019 Bo-Yi Wu"
	app.Authors = []cli.Author{
		{
			Name:  "Bo-Yi Wu",
			Email: "appleboy.tw@gmail.com",
		},
	}
	app.Action = run
	app.Version = Version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "host",
			Usage:  "jenkins base url",
			EnvVar: "PLUGIN_URL,JENKINS_URL,INPUT_URL",
		},
		cli.StringFlag{
			Name:   "user,u",
			Usage:  "jenkins username",
			EnvVar: "PLUGIN_USER,JENKINS_USER,INPUT_USER",
		},
		cli.StringFlag{
			Name:   "token,t",
			Usage:  "jenkins token",
			EnvVar: "PLUGIN_TOKEN,JENKINS_TOKEN,INPUT_TOKEN",
		},
		cli.StringSliceFlag{
			Name:   "job,j",
			Usage:  "jenkins job",
			EnvVar: "PLUGIN_JOB,JENKINS_JOB,INPUT_JOB",
		},
		cli.BoolFlag{
			Name:   "insecure",
			Usage:  "allow insecure server connections when using SSL",
			EnvVar: "PLUGIN_INSECURE,JENKINS_INSECURE,INPUT_INSECURE",
		},
		cli.StringSliceFlag{
			Name:   "parameters,p",
			Usage:  "jenkins build parameters",
			EnvVar: "PLUGIN_PARAMETERS,JENKINS_PARAMETERS,INPUT_PARAMETERS",
		},
	}

	// Override a template
	cli.AppHelpTemplate = `
________                                            ____.              __   .__
\______ \_______  ____   ____   ____               |    | ____   ____ |  | _|__| ____   ______
 |    |  \_  __ \/  _ \ /    \_/ __ \   ______     |    |/ __ \ /    \|  |/ /  |/    \ /  ___/
 |    |   \  | \(  <_> )   |  \  ___/  /_____/ /\__|    \  ___/|   |  \    <|  |   |  \\___ \
/_______  /__|   \____/|___|  /\___  >         \________|\___  >___|  /__|_ \__|___|  /____  >
        \/                  \/     \/                        \/     \/     \/       \/     \/
                                                                    version: {{.Version}}
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
    Github: https://github.com/appleboy/drone-line
`

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	plugin := Plugin{
		BaseURL:    c.String("host"),
		Username:   c.String("user"),
		Token:      c.String("token"),
		Job:        c.StringSlice("job"),
		Insecure:   c.Bool("insecure"),
		Parameters: c.StringSlice("parameters"),
	}

	return plugin.Exec()
}
