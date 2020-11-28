# mqtt-osc

[![GoDoc](https://godoc.org/github.com/72nd/mqtt-osc?status.svg)](https://godoc.org/github.com/72nd/mqtt-osc) [![Go Report Card](https://goreportcard.com/badge/github.com/72nd/mqtt-osc)](https://goreportcard.com/report/github.com/72nd/mqtt-osc) ![GitHub](https://img.shields.io/github/license/72nd/mqtt-osc)

<p align="center">
  <img width="512" src="misc/logo.png">
</p>

This library provides a mechanism to issue [Open Sound Control](https://en.wikipedia.org/wiki/Open_Sound_Control) messages on [MQTT](https://mqtt.org/) events. Currently the flow of information is only supported from MQTT to OSC. There is also a simple command line application providing the basic relaying of messages. Using the library provides much more power of control.

## Usage CLI

There is a simple command line interface for this library. You can download the binary for your platform in the Release section. Configuration is done via a [YAML](https://en.wikipedia.org/wiki/YAML) file. The tool provides you with a command to create a initial configuration file. Just do:

```shell script
mqtt-osc config config.yaml
``` 

This will create the file `config.yaml`. Then open this file in your favorite text editor and tweak it for your usage.

```yaml
# Hostname of the MQTT broker (aka server).
mqtt_host: 127.0.0.1
# Port of the MQTT broker (aka server).
mqtt_port: 1883
# Sets the Client-ID when connecting to the MQTT broker.
mqtt_client_id: mqtt-osc-relay
# Username for the authentication against the MQTT broker.
mqtt_user: user
# Password for the authentication against the MQTT broker.
mqtt_password: secret
# Hostname of the OSC server where the OSC command should be sent to.
osc_host: 127.0.0.1
# Port of the OSC server where the OSC command should be sent to.
osc_port: 8765
# A list of handlers. Each handler defines a MQTT topic which will be
# trigerred by. For More information see below.
handlers:
- mqtt_topic: /light/+/on
  osc_address: /light/{{ ._1 }}/turn-on
  # If true, the payload of the MQTT event will be relayed to the OSC
  # server as a string content.
  relay_payload: false
```

