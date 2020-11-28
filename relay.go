package mqttosc

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"

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

// Hanlder listen to one MQTT event and describes the effects it will
// have. This library contains some handler functions for the
// most common cases.
//
// To increase the flexibility of the handlers the output OSC Address
// can contain template items. The Translate function can be used to
// define the data for this template. The matching-groups from the MQTT
// topic can be accessed by index using `{{ ._1 }}`. The content of the
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
	// no keys beginning with `_` (ex `_1`) are allowed in this map as these
	// are used to store the content of the wildcard match-groups.
	Translage *func(topic string, data string) map[string]string `yaml:"-"`
	// RelayPayload states whether the MQTT message should be relayed to
	// the OSC recipient.
	RelayPayload bool `yaml:"relay_payload"`
	// logFunc takes the logger function of the Handler's relay.
	logFunc Logger `yaml:"-"`
	// topicRegex is the regular expression used to capture the wildcards as
	// capture groups.
	topicRegex *regexp.Regexp
	// oscAddressTemplate is the template of the OSC address. The template is
	// instantiated on initialization to safe some time on event trigger.
	oscAddressTemplate *template.Template
}

// init has to be called before using or registering the handler to the MQTT
// client. The method sets the log function, determines the correct topicRegex
// based on the MqttTopic and
func (h *Handler) init(logFunc Logger) error {
	h.logFunc = logFunc
	h.topicRegex = regexForTopic(h.MqttTopic)

	fmt.Println(h.OscAddress)

	tpl, err := template.New(h.MqttTopic).Parse(h.OscAddress)
	if err != nil {
		return err
	}
	h.oscAddressTemplate = tpl
	return nil
}

// onEvent is internally called when the MQTT topic was updated.
func (h Handler) onEvent(client mqtt.Client, message mqtt.Message) {
	h.logFunc(DebugLevel, "handler \"%s\" was triggered by message on topic \"%s\"", h.MqttTopic, message.Topic())
	h.logFunc(DebugLevel, "lala: %s", h.topicRegex)
}

// regexForTopic converts a given address for an MQTT topic to a regex to
// match the content of the wildcards against it.
func regexForTopic(topic string) *regexp.Regexp {
	parts := strings.Split(topic, "/")
	rsl := make([]string, len(parts))
	for i, part := range parts {
		if part == "+" || part == "*" {
			rsl[i] = "(.*)"
		} else {
			rsl[i] = part
		}
	}
	return regexp.MustCompile(strings.Join(rsl, "/"))
}
