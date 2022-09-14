package main

import (
	"fmt"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/energomonitor/bisquitt"
	"github.com/energomonitor/bisquitt/util/platform"
)

const (
	MqttHostFlag             = "mqtt-host"
	MqttPortFlag             = "mqtt-port"
	MqttUserFlag             = "mqtt-user"
	MqttPasswordFlag         = "mqtt-password"
	MqttPasswordFileFlag     = "mqtt-password-file"
	MqttTimeoutFlag          = "mqtt-timeout"
	HostFlag                 = "host"
	PortFlag                 = "port"
	DtlsFlag                 = "dtls"
	SelfSignedFlag           = "self-signed"
	CertFlag                 = "cert"
	KeyFlag                  = "key"
	PredefinedTopicFlag      = "predefined-topic"
	PredefinedTopicsFileFlag = "predefined-topics-file"
	SyslogFlag               = "syslog"
	DebugFlag                = "debug"
	PerformanceLogTimeFlag   = "performance-log-time"
	InsecureFlag             = "insecure"
	AuthFlag                 = "auth"
	UserFlag                 = "user"
	GroupFlag                = "group"
)

var Application = cli.App{
	Name:        "bisquitt",
	Usage:       "A transparent MQTT-SN gateway with DTLS support",
	ArgsUsage:   " ",
	Version:     bisquitt.Version(),
	Description: "A transparent MQTT-SN gateway with DTLS support.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  MqttHostFlag,
			Usage: "MQTT broker host",
			Value: "127.0.0.1",
			EnvVars: []string{
				"MQTT_HOST",
			},
		},
		&cli.IntFlag{
			Name:  MqttPortFlag,
			Usage: "MQTT broker port",
			Value: 1883,
			EnvVars: []string{
				"MQTT_PORT",
			},
		},
		&cli.StringFlag{
			Name:  MqttUserFlag,
			Usage: "username for MQTT broker connection",
			EnvVars: []string{
				"MQTT_USER",
			},
		},
		&cli.StringFlag{
			Name:  MqttPasswordFlag,
			Usage: "password for MQTT broker connection",
			EnvVars: []string{
				"MQTT_PASSWORD",
			},
		},
		&cli.PathFlag{
			Name:  MqttPasswordFileFlag,
			Usage: "password file for MQTT broker connection",
			EnvVars: []string{
				"MQTT_PASSWORD_FILE",
			},
		},
		&cli.DurationFlag{
			Name:  MqttTimeoutFlag,
			Usage: "MQTT connection timeout",
			Value: 30 * time.Second,
			EnvVars: []string{
				"MQTT_TIMEOUT",
			},
		},
		&cli.StringFlag{
			Name:  HostFlag,
			Usage: "host to listen on",
			Value: "localhost",
			EnvVars: []string{
				"HOST",
			},
		},
		&cli.IntFlag{
			Name:  PortFlag,
			Usage: "port to listen on",
			Value: 1883,
			EnvVars: []string{
				"PORT",
			},
		},
		&cli.BoolFlag{
			Name:  DtlsFlag,
			Usage: "use DTLS",
			EnvVars: []string{
				"DTLS_ENABLED",
			},
		},
		&cli.BoolFlag{
			Name:  SelfSignedFlag,
			Usage: "generate self-signed certificate",
			EnvVars: []string{
				"SELF_SIGNED",
			},
		},
		&cli.PathFlag{
			Name:  CertFlag,
			Usage: "DTLS certificate file",
			EnvVars: []string{
				"CERT_FILE",
			},
		},
		&cli.PathFlag{
			Name:  KeyFlag,
			Usage: "DTLS key file",
			EnvVars: []string{
				"KEY_FILE",
			},
		},
		&cli.StringSliceFlag{
			Name:  PredefinedTopicFlag,
			Usage: fmt.Sprintf("predefined topic, takes precedence over --%s (format: clientID;topicName;topicID)", PredefinedTopicsFileFlag),
			EnvVars: []string{
				"PREDEFINED_TOPIC",
			},
		},
		&cli.PathFlag{
			Name:  PredefinedTopicsFileFlag,
			Usage: "file with pre-defined topics",
			EnvVars: []string{
				"PREDEFINED_TOPICS_FILE",
			},
		},
		&cli.BoolFlag{
			Name:  SyslogFlag,
			Usage: "log to syslog",
			EnvVars: []string{
				"SYSLOG",
			},
		},
		&cli.BoolFlag{
			Name:  DebugFlag,
			Usage: "print debug messages",
			EnvVars: []string{
				"DEBUG",
			},
		},
		&cli.DurationFlag{
			Name:  PerformanceLogTimeFlag,
			Usage: "performance log frequency",
			Value: 0,
			EnvVars: []string{
				"PERFORMANCE_LOG_TIME",
			},
		},
		&cli.BoolFlag{
			Name:  InsecureFlag,
			Usage: "allow plaintext authentication over unencrypted channel",
			EnvVars: []string{
				"INSECURE",
			},
		},
		&cli.BoolFlag{
			Name:  AuthFlag,
			Usage: "enable non-standard AUTH support",
			Value: false,
			EnvVars: []string{
				"AUTH",
			},
		},
		&cli.StringFlag{
			Name:  UserFlag,
			Usage: "run gateway as a user",
			Value: "",
			EnvVars: []string{
				"BISQUITT_USER",
			},
			Hidden: !platform.HasSetUser(),
		},
		&cli.StringFlag{
			Name:  GroupFlag,
			Usage: "run gateway as a group",
			Value: "",
			EnvVars: []string{
				"BISQUITT_GROUP",
			},
			Hidden: !platform.HasSetGroup(),
		},
	},
	HideHelpCommand: true,
	Action:          handleAction(),
}
