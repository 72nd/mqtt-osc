package mqttosc

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type LogLevel int

const (
	LogTrace LogLevel = iota
	LogDebug
	LogInfo
	LogWarn
	LogError
	LogPanic
)

func (l LogLevel) String() string {
	return [...]string{"Trace", "Debug", "Info", "Warn", "Error"}[l]
}

// Logger is the log function used in the library. It takes the message
// and defines the handling of this message.
type Logger func(msg string, level LogLevel)

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
	// Handlers is the collection of handler the relay handles.
	Handlers []Handler `yaml:"handlers"`
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
		r.log(fmt.Sprintf("MQTT error, %s", token.Error()), LogError)
		return
	}

	for i := range r.Handlers {
		r.Handlers[i].logFunc = *r.LogFunc
		token := client.Subscribe(r.Handlers[i].MqttTopic, 1, r.Handlers[i].onEvent)
		token.Wait()
	}

	for {
	}
}

// log a message using the in the LogFunc defined log method. If
// no method is provided, the output is written to the standard
// output.
func (r Relay) log(msg string, level LogLevel) {
	if r.LogFunc != nil {
		var fn Logger
		fn = *r.LogFunc
		fn(msg, level)
	} else {
		if level == LogPanic {
			panic(msg)
		}
		fmt.Printf("%s: %s", level, msg)
	}
}

// Hanlder listen to one MQTT event and describes the effects it will
// have. This library contains some handler functions for the
// most common cases.
//
// To increase the flexibility of the handlers the output OSC Address
// can contain template items. The Translate function can be used to
// define the data for this template. The matching-groups from the MQTT
// topic can be accessed by index using `{{ .$1 }}`. The content of the
// MQTT payload (also known as message) can be accessed using the
// `{{ .Payload }}` argument.
type Handler struct {
	// Path to the MQTT topic to be listen to.
	MqttTopic string `yaml:"mqtt_topic"`
	// OscAddress points to the address the OSC command should be sent
	// on an MQTT event. Learn more about the ability to apply templates
	// in the Handler type documentation.
	OscAddress string `yaml:"osc_address"`
	// Translate is the function which can be used to alter the output
	// OSC address and payload. It gets the concrete MQTT topic which was
	// called as well as the payload (if any). It returns a map of strings
	// which will be applied to the template in the OSC address. Note that
	// no keys beginning with `$` (ex `$1`) are allowed in this map as these
	// are used to store the content of the wildcard match-groups.
	Translage *func(topic string, data string) map[string]string `yaml:"-"`
	// RelayPayload states whether the MQTT message should be relayed to
	// the OSC recipient.
	RelayPayload bool `yaml:"relay_payload"`
	// logFunc takes the logger function of the Handler's relay.
	logFunc Logger `yaml:"-"`
}

// onEvent is internally called when the MQTT topic was updated.
func (h Handler) onEvent(client mqtt.Client, message mqtt.Message) {
}
