package main

import (
	"io/ioutil"
	"os"

	"github.com/72nd/mqtt-osc"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

func main() {
	app := &cli.App{
		Name:  "mqtt-osc",
		Usage: "relaying updates on MQTT topics to OSC",
		Action: func(c *cli.Context) error {
			_ = cli.ShowCommandHelp(c, c.Command.Name)
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "config",
				Aliases: []string{"cfg"},
				Usage:   "generate a new configuration file",
				Action: func(c *cli.Context) error {
					createConfig(getPath(c))
					return nil
				},
			},
			{
				Name:    "run",
				Aliases: []string{"serve"},
				Usage:   "run the relay",
				Action: func(c *cli.Context) error {
					if c.Bool("debug") {
						logrus.SetLevel(logrus.DebugLevel)
					}
					run(getPath(c))
					return nil
				},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "debug",
						Aliases: []string{"d"},
						Usage:   "enable debug mode",
					},
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal("cli app error: ", err)
	}
}

// getPath tries to get the first positional argument given by the user.
// Or fatals with a error message.
func getPath(c *cli.Context) string {
	if c.Args().Len() != 1 {
		logrus.Fatal("one positional argument needed (path to config file)")
	}
	return c.Args().Get(0)
}

// createConfig creates a new config file with default values and saves it to
// the given path.
func createConfig(path string) {
	relay := mqttosc.Relay{
		MqttHost:     "127.0.0.1",
		MqttPort:     1883,
		MqttClientId: "mqtt-osc-relay",
		MqttUser:     "user",
		MqttPassword: "secret",
		Handlers: []mqttosc.Handler{
			{
				MqttTopic:  "/light/+/on",
				OscAddress: "/light/{{ .$1 }}/turn-on",
			},
		},
	}
	data, err := yaml.Marshal(relay)
	if err != nil {
		logrus.Fatalf("couldn't marshal config, %s", err)
	}
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		logrus.Fatalf("couldn't write config to %s, %s", path, data)
	}
}

// run runs the relay.
func run(path string) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		logrus.Fatalf("error loading configuration form %s, %s", path, err)
	}
	var relay mqttosc.Relay
	if err := yaml.Unmarshal(raw, &relay); err != nil {
		logrus.Fatalf("error unmarshalling configuration from %s, %s", path, err)
	}

	var logger mqttosc.Logger
	logger = logFunc
	relay.LogFunc = &logger

	logrus.Infof("listen to MQTT broker %s:%d", relay.MqttHost, relay.MqttPort)
	relay.Serve()
}

// logFunc defines the Logger function for the mqttosc.Relay using logrus.
func logFunc(level mqttosc.LogLevel, format string, args ...interface{}) {
	switch level {
	case mqttosc.TraceLevel:
		logrus.Tracef(format, args...)
	case mqttosc.DebugLevel:
		logrus.Debugf(format, args...)
	case mqttosc.InfoLevel:
		logrus.Infof(format, args...)
	case mqttosc.WarnLevel:
		logrus.Warnf(format, args...)
	case mqttosc.ErrorLevel:
		logrus.Errorf(format, args...)
	case mqttosc.PanicLevel:
		logrus.Panicf(format, args...)
	}
}
