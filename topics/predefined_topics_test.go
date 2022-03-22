package topics

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPredefinedTopics_Add(t *testing.T) {
	const (
		clientID  = "client1"
		topicName = "device/any/data"
		topicID   = uint16(1)
	)

	expected := PredefinedTopics{
		clientID: {
			topicID: topicName,
		},
	}
	actual := PredefinedTopics{}
	actual.Add(clientID, topicName, topicID)
	assert.Equal(t, expected, actual)
}

var clientID1 = "client1"
var topics1 = PredefinedTopics{
	clientID1: map[uint16]string{
		1: "device/000001/data",
		2: "device/000001/config",
	},
	"*": map[uint16]string{
		1: "device/any/data",
		2: "device/any/config",
		3: "device/any/bcast",
	},
}

func TestPredefinedTopics_GetTopicName(t *testing.T) {
	assert := assert.New(t)

	topic, ok := topics1.GetTopicName(clientID1, 1)
	assert.True(ok)
	assert.Equal("device/000001/data", topic)

	topic, ok = topics1.GetTopicName(clientID1, 2)
	assert.True(ok)
	assert.Equal("device/000001/config", topic)

	topic, ok = topics1.GetTopicName(clientID1, 3)
	assert.True(ok)
	assert.Equal("device/any/bcast", topic)

	topic, ok = topics1.GetTopicName(clientID1, 4)
	assert.False(ok)
}

func TestPredefinedTopics_GetTopicID(t *testing.T) {
	assert := assert.New(t)

	topicID, ok := topics1.GetTopicID(clientID1, "device/000001/data")
	assert.True(ok)
	assert.Equal(uint16(1), topicID)

	topicID, ok = topics1.GetTopicID(clientID1, "device/000001/config")
	assert.True(ok)
	assert.Equal(uint16(2), topicID)

	topicID, ok = topics1.GetTopicID(clientID1, "device/any/bcast")
	assert.True(ok)
	assert.Equal(uint16(3), topicID)

	topicID, ok = topics1.GetTopicID(clientID1, "nonexistent")
	assert.False(ok)
}

func TestMergePredefinedTopics(t *testing.T) {
	src := PredefinedTopics{
		"client1": {
			1: "src/client1/1",
			2: "src/client1/2",
		},
		"client2": {
			1: "src/client2/1",
		},
		"*": {
			1: "src/*/1",
			2: "src/*/2",
		},
	}
	dest := PredefinedTopics{
		"client1": {
			1: "dest/client1/1",
			3: "dest/client1/3",
		},
		"client3": {
			1: "dest/client3/1",
		},
		"*": {
			1: "dest/*/1",
			3: "dest/*/3",
		},
	}
	expected := PredefinedTopics{
		"client1": {
			1: "src/client1/1",
			2: "src/client1/2",
			3: "dest/client1/3",
		},
		"client2": {
			1: "src/client2/1",
		},
		"client3": {
			1: "dest/client3/1",
		},
		"*": {
			1: "src/*/1",
			2: "src/*/2",
			3: "dest/*/3",
		},
	}
	dest.Merge(src)
	assert.Equal(t, expected, dest)
}

func TestParsePredefinedTopicOptions(t *testing.T) {
	t.Run("With client ID", func(t *testing.T) {
		expected := PredefinedTopics{
			"client1": map[uint16]string{
				1: "messages",
			},
		}
		actual, err := ParsePredefinedTopicOptions("client1;messages;1")
		assert.Nil(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("Without client ID", func(t *testing.T) {
		expected := PredefinedTopics{
			"*": map[uint16]string{
				1: "messages",
			},
		}
		actual, err := ParsePredefinedTopicOptions("messages;1")
		assert.Nil(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("With invalid order", func(t *testing.T) {
		_, err := ParsePredefinedTopicOptions("client1;1;messages")
		assert.Error(t, err)

		_, err = ParsePredefinedTopicOptions("1;messages")
		assert.Error(t, err)
	})

	t.Run("With multiple lines", func(t *testing.T) {
		lines := []string{
			"client1;device/000001/data;1",
			"client1;device/000001/config;2",
			"*;device/any/data;1",
			"*;device/any/config;2",
			"*;device/any/bcast;3",
		}
		expected := PredefinedTopics{
			"client1": {
				1: "device/000001/data",
				2: "device/000001/config",
			},
			"*": {
				1: "device/any/data",
				2: "device/any/config",
				3: "device/any/bcast",
			},
		}
		actual, err := ParsePredefinedTopicOptions(lines...)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expected, actual)
	})
}

func TestReadPredefinedTopicsFile(t *testing.T) {
	expected := PredefinedTopics{
		"client1": {
			1: "device/000001/data",
			2: "device/000001/config",
		},
		"*": {
			1: "device/any/data",
			2: "device/any/config",
			3: "device/any/bcast",
		},
	}
	actual, err := ReadPredefinedTopicsFile("testdata/topics.yaml")
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}

func ExampleParsePredefinedTopicOptions() {
	option1 := "*;test/topic1;1"
	option2 := "client;test/topic2;2"

	topics, _ := ParsePredefinedTopicOptions(option1, option2)

	fmt.Println(topics)

	// Output:
	// map[*:map[1:test/topic1] client:map[2:test/topic2]]
}

func ExampleReadPredefinedTopicsFile() {
	f, _ := ioutil.TempFile("/tmp", "example_predefined_topics.yaml")
	defer os.Remove(f.Name())

	f.Write([]byte(`
"*":
  1: any/topic1
  2: any/topic2
client1:
  3: client1/topic1
  4: client1/topic2`))
	f.Close()

	topics, _ := ReadPredefinedTopicsFile(f.Name())
	fmt.Println(topics)

	// Output:
	// map[*:map[1:any/topic1 2:any/topic2] client1:map[3:client1/topic1 4:client1/topic2]]
}
