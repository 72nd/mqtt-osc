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
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "config",
				Usage: "generate a new configuration file",
				Action: func(c *cli.Context) error {
					return nil
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal("cli app error: ", err)
	}
}

// getArgument tries to get the first positional argument given by the user.
// Or fatals with a error message.
func getArgument(c *cli.Context) string {
	if c.Args().Len() != 1 {
		logrus.Fatalf("one positional argument needed (path to config file)")
	}
	return c.Args().Get(0)
}

// createConfig creates a new config file with default values and saves it to
// the given path.
func createConfig(path string) {
	var logger mqttosc.Logger
	logger = logFunc
	relay := mqttosc.Relay{
		MqttHost:     "127.0.0.1",
		MqttPort:     6379,
		MqttClientId: "mqtt-osc-relay",
		MqttUser:     "user",
		MqttPassword: "secret",
		Handlers: []mqttosc.Handler{
			{
				MqttTopic:  "/light/+/on",
				OscAddress: "/light/{{ .$1 }}/turn-on",
			},
		},
		LogFunc: &logger,
	}
	data, err := yaml.Marshal(relay)
	if err != nil {
		logrus.Fatalf("couldn't marshal config, %s", err)
	}
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		logrus.Fatalf("couldn't write config to %s, %s", path, data)
	}
}

// logFunc defines the Logger function for the mqttosc.Relay using logrus.
func logFunc(msg string, level mqttosc.LogLevel) {
	switch level {
	case mqttosc.LogTrace:
		logrus.Trace(msg)
	case mqttosc.LogDebug:
		logrus.Debug(msg)
	case mqttosc.LogInfo:
		logrus.Info(msg)
	case mqttosc.LogWarn:
		logrus.Warn(msg)
	case mqttosc.LogError:
		logrus.Error(msg)
	}
}
