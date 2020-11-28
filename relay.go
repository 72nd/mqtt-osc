package mqttosc

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type LogLevel int

const (
	TraceLevel LogLevel = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	PanicLevel
)

func (l LogLevel) String() string {
	return [...]string{"Trace", "Debug", "Info", "Warn", "Error"}[l]
}

// Logger is the log function used in the library. It takes the message
// and defines the handling of this message. The message can be used as
// format string.
type Logger func(level LogLevel, format string, args ...interface{})

// Relay provides the forwarding of messages from MQTT to OSC. It consist
// contains a slice of Handlers each describing one MQTT event to be
// listen to.
type Relay struct {
	// MqttHost is the hostname of the MQTT broker.
	MqttHost string `yaml:"mqtt_host"`
	// MqttPort is the port of the MQTT broker.
	MqttPort int `yaml:"mqtt_port"`
	// MqttClientId sets the MQTT Id of this Relay.
	MqttClientId string `yaml:"mqtt_client_id"`
	// MqttUser states the user name used for the authentication
	// with the MQTT broker.
	MqttUser string `yaml:"mqtt_user"`
	// MqttUser states the password used for the authentication
	// with the MQTT broker.
	MqttPassword string `yaml:"mqtt_password"`
	// OscHost is the hostname of the OSC server.
	OscHost string `yaml:"osc_host"`
	// OscPort is the port of the OSC server.
	OscPort int `yaml:"osc_port"`
	// Handlers is the collection of handler the relay handles.
	Handlers []MqttToOscHandler `yaml:"handlers"`
	// LogFunc provides the possibility to customize the log functionality.
	// The function is called on each log. If no method is set, the debug
	// output will be outputted to standard output.
	LogFunc *Logger `yaml:"-"`
}

// Serve starts the MQTT client and waits for incoming updates on the topics
// define by the handlers.
func (r *Relay) Serve() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", r.MqttHost, r.MqttPort))
	opts.SetClientID(r.MqttClientId)
	opts.SetUsername(r.MqttUser)
	opts.SetPassword(r.MqttPassword)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		r.log(ErrorLevel, "MQTT error, %s", token.Error())
		return
	}

	for i := range r.Handlers {
		if err := r.Handlers[i].init(*r.LogFunc); err != nil {
			r.log(ErrorLevel, "couldn't initialize %s, %s", r.Handlers[i].MqttTopic, err)
			return
		}
		token := client.Subscribe(r.Handlers[i].MqttTopic, 1, r.Handlers[i].onEvent)
		token.Wait()
	}

	for {
	}
}

// log a message using the in the LogFunc defined log method. If
// no method is provided, the output is written to the standard
// output.
func (r Relay) log(level LogLevel, format string, args ...interface{}) {
	if r.LogFunc != nil {
		var fn Logger
		fn = *r.LogFunc
		fn(level, format, args...)
	} else {
		if level == PanicLevel {
			panic(fmt.Sprintf(format, args...))
		}
		fmt.Printf("%s: %s", level, fmt.Sprintf(format, args...))
	}
}
