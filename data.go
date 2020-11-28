package mqttosc

import "fmt"

// Logger is the log function used in the library. It takes the message
// and defines the handling of this message.
type Logger func(msg string)

// Relay provides the forwarding of messages from MQTT to OSC. It consist
// contains a slice of Handlers each describing one MQTT event to be
// listen to.
type Relay struct {
	// MqttHost is the hostname of the MQTT broker.
	MqttHost string
	// MqttPort is the port of the MQTT broker.
	MqttPort int
	// MqttClientId sets the MQTT Id of this Relay.
	MqttClientId string
	// MqttUser states the user name used for the authentication
	// with the MQTT broker.
	MqttUser string
	// MqttUser states the password used for the authentication
	// with the MQTT broker.
	MqttPassword string
	// Handlers is the collection of handler the relay handles.
	Handlers []Handler
	// LogFunc provides the possibility to customize the log
	// functionality. The function is called on each debug log.
	// If no method is set, the debug output will be outputted to
	// standard output.
	LogFunc *Logger
}

// log a message using the in the LogFunc defined log method. If
// no method is provided, the output is written to the standard
// output.
func (r Relay) log(msg string) {
	if r.LogFunc != nil {
		var fn Logger
		fn = *r.LogFunc
		fn(msg)
	} else {
		fmt.Println(msg)
	}
}

//

// Hanlder listen to one MQTT event and describes the effects it will
// have. This library contains some handler functions for the
// most common cases.
// 
// To increase the flexibility of the handlers the output OSC Address
// can contain template items. Per default 
// The string can contain references to match-groups
// defined trough wildcards in the MQTT address. Example
type Handler struct {
	// Path to the MQTT topic to be listen to.
	MqttTopic string
	// OscAddress points to the address the OSC command should be sent
	// on an MQTT event. Learn more about the ability to apply templates
	// in the Handler type documentation.
	OscAddress string
	// Translate is the function which can be used to alter the output
	// OSC address and payload. It gets the concrete MQTT topic which was
	// called as well as the payload (if any). It returns a map of strings
	// which will be applied to the template in the OSC address. Note that
	// no keys beginning with `$` (ex `$1`) are allowed in this map as these
	// are used to store the content of the wildcard match-groups. 
	Translage func(topic string, data string) map[string]string
}
