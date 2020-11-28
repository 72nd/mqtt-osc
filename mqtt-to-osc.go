package mqttosc

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// TranslateFunc is the function type used to alter the outgoing OSC
// address template data and payload.
type TranslateFunc func(topic string, data string) (tplData map[string]interface{}, payload string)

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
type MqttToOscHandler struct {
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
	Translate *TranslateFunc `yaml:"-"`
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
// based on the MqttTopic and. This method checks also if the Translate function
// introduces any illegal keys (everything starting with `_`).
func (h *MqttToOscHandler) init(logFunc Logger) error {
	h.logFunc = logFunc
	h.topicRegex = regexForTopic(h.MqttTopic)

	fmt.Println(h.OscAddress)

	tpl, err := template.New(h.MqttTopic).Parse(h.OscAddress)
	if err != nil {
		return err
	}
	h.oscAddressTemplate = tpl

	if err := h.checkTranslate(); err != nil {
		return err
	}

	return nil
}

// checkTranslate checks the output of the translate method given by the library
// user. Will return an error if there is a template data key starting with an
// underscore (`_`).
func (h MqttToOscHandler) checkTranslate() error {
	if h.Translate == nil {
		return nil
	}
	var fn TranslateFunc
	fn = *h.Translate
	userData, _ := fn("", "")
	for key, _ := range userData {
		if strings.HasPrefix(key, "_") {
			return fmt.Errorf("it's forbidden to introduce map elements with keys starting with '_'")
		}
	}
	return nil
}

// onEvent is internally called when the MQTT topic was updated.
func (h MqttToOscHandler) onEvent(client mqtt.Client, message mqtt.Message) {
	h.logFunc(DebugLevel, "handler \"%s\" was triggered by message on topic \"%s\"", h.MqttTopic, message.Topic())

	tplData := make(map[string]interface{}, 1)
	// var payload string
	if h.Translate != nil {
		var fn TranslateFunc
		fn = *h.Translate
		tplData, _ = fn(message.Topic(), string(message.Payload()))
	}
	tplData = h.templateData(tplData, message.Topic())
	var adr bytes.Buffer
	if err := h.oscAddressTemplate.Execute(&adr, tplData); err != nil {
		h.logFunc(ErrorLevel, "failed to execute template for OSC address %s, %s", h.OscAddress, err)
	}
}

// templateData returns the map to apply on the oscAddressTemplate. It parses the
// concrete content of any wildcards and adds this to the user chosen data from the
// translate method. Topic parameter needs the concrete topic of the incoming MQTT
// event.
func (h MqttToOscHandler) templateData(userData map[string]interface{}, topic string) map[string]interface{} {
	matches := h.topicRegex.FindStringSubmatch(topic)
	for i := range matches {
		if i == 0 {
			continue
		}
		userData[fmt.Sprintf("_%d", i)] = matches[i]
	}
	return userData
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
