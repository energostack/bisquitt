package client

import (
	"fmt"
	"strings"
	"sync"

	msgs "github.com/energomonitor/bisquitt/messages"
)

// Subscribed message handler callback type.
type MessageHandlerFunc func(client *Client, topic string, msg *msgs.PublishMessage)

type messageHandler struct {
	route    []string
	callback MessageHandlerFunc
}

type messageHandlers struct {
	handlers       sync.Map
	defaultHandler MessageHandlerFunc
}

func (mhs *messageHandlers) add(route []string, callback MessageHandlerFunc) {
	mhs.handlers.Store(join(route), &messageHandler{
		route:    route,
		callback: callback,
	})
}

func (mhs *messageHandlers) handle(client *Client, topic string, msg *msgs.PublishMessage) {
	var callback MessageHandlerFunc
	route := split(topic)
	mhs.handlers.Range(func(key, value interface{}) bool {
		mh, ok := value.(*messageHandler)
		if !ok {
			panic(fmt.Errorf("unexpected type '%T'", value))
		}

		if match(mh.route, route) {
			callback = mh.callback
			return false
		}

		return true
	})

	if callback != nil {
		go callback(client, topic, msg)
		return
	} else {
		if mhs.defaultHandler != nil {
			go mhs.defaultHandler(client, topic, msg)
			return
		}
	}
}

func join(route []string) string {
	return strings.Join(route, "/")
}

func split(topic string) []string {
	return strings.Split(topic, "/")
}

// Taken from Paho mqtt client:
// https://github.com/eclipse/paho.mqtt.golang/blob/a140ed81404c0a4aa0e97c91e7b99d1577c45418/router.go#L33
//
// match takes a slice of strings which represent the route being tested having been split on '/'
// separators, and a slice of strings representing the topic string in the published message, similarly
// split.
// The function determines if the topic string matches the route according to the MQTT topic rules
// and returns a boolean of the outcome
func match(route []string, topic []string) bool {
	if len(route) == 0 {
		return len(topic) == 0
	}

	if len(topic) == 0 {
		return route[0] == "#"
	}

	if route[0] == "#" {
		return true
	}

	if (route[0] == "+") || (route[0] == topic[0]) {
		return match(route[1:], topic[1:])
	}

	return false
}
