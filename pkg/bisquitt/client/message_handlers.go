package client

import (
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
	sync.RWMutex
	handlers       []*messageHandler
	defaultHandler MessageHandlerFunc
}

func (mhs *messageHandlers) add(route []string, callback MessageHandlerFunc) {
	mhs.Lock()
	defer mhs.Unlock()

	mhs.handlers = append(mhs.handlers, &messageHandler{
		route:    route,
		callback: callback,
	})
}

func (mhs *messageHandlers) setDefaultHandler(callback MessageHandlerFunc) {
	mhs.Lock()
	defer mhs.Unlock()

	mhs.defaultHandler = callback
}

func (mhs *messageHandlers) handle(client *Client, topic string, msg *msgs.PublishMessage) {
	mhs.RLock()
	defer mhs.RUnlock()

	for _, mh := range mhs.handlers {
		if match(mh.route, strings.Split(topic, "/")) {
			go mh.callback(client, topic, msg)
			return
		}
	}

	if mhs.defaultHandler != nil {
		go mhs.defaultHandler(client, topic, msg)
	}
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
