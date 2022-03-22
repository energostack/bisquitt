// Package topics contains tools related to MQTT-SN topics handling.
package topics

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// PredefinedTopics implements a map of MQTT-SN predefined topics.
// It maps clientID -> (topicID -> topicName).
type PredefinedTopics map[string]map[uint16]string

// Add adds a new predefined topic to the map.
func (t PredefinedTopics) Add(clientID, topicName string, topicID uint16) {
	if _, ok := t[clientID]; !ok {
		t[clientID] = map[uint16]string{}
	}
	t[clientID][topicID] = topicName
}

// GetTopicName returns a topic name for the given clientID and topicID.
func (t PredefinedTopics) GetTopicName(clientID string, topicID uint16) (string, bool) {
	if tClient, ok := t[clientID]; ok {
		if topicName, ok := tClient[topicID]; ok {
			return topicName, true
		}
	}
	if tAll, ok := t["*"]; ok {
		if topicName, ok := tAll[topicID]; ok {
			return topicName, true
		}
	}
	return "", false
}

// GetTopicID returns a topic ID for the given clientID and topic.
func (t PredefinedTopics) GetTopicID(clientID, topic string) (uint16, bool) {
	if tClient, ok := t[clientID]; ok {
		for topicID, topicName := range tClient {
			if topicName == topic {
				return topicID, true
			}
		}
	}
	if tAll, ok := t["*"]; ok {
		for topicID, topicName := range tAll {
			if topicName == topic {
				return topicID, true
			}
		}
	}
	return 0, false
}

// Merge merges another predefined topics map in. If there are collisions,
// the given map (src) values take precedence.
func (t PredefinedTopics) Merge(src PredefinedTopics) {
	for clientID := range src {
		if _, ok := t[clientID]; !ok {
			t[clientID] = src[clientID]
			continue
		}
		for topicID := range src[clientID] {
			t[clientID][topicID] = src[clientID][topicID]
		}
	}
}

// ReadPredefinedTopicsFile reads a predefined topics definition file in YAML format.
func ReadPredefinedTopicsFile(file string) (PredefinedTopics, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := PredefinedTopics{}
	err = yaml.NewDecoder(f).Decode(result)
	return result, err
}

// ParsePredefinedTopicOptions parses a command line predefined topics definition
// in "client_id;topic;topic_id" format.
func ParsePredefinedTopicOptions(options ...string) (PredefinedTopics, error) {
	parseLine := func(fields []string) (clientID string, topicName string, topicID uint64, err error) {
		switch len(fields) {
		case 2:
			clientID = "*"
			topicName = fields[0]
			topicID, err = strconv.ParseUint(fields[1], 10, 16)
			return
		case 3:
			clientID = fields[0]
			topicName = fields[1]
			topicID, err = strconv.ParseUint(fields[2], 10, 16)
			return
		default:
			err = errors.New("invalid format (expects: clientID;topicName;topicID)")
			return
		}
	}

	result := make(PredefinedTopics)
	for _, line := range options {
		clientID, topicName, topicID, err := parseLine(strings.Split(line, ";"))
		if err != nil {
			return nil, err
		}
		result.Add(clientID, topicName, uint16(topicID))
	}
	return result, nil
}
