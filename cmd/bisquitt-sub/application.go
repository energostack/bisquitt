package main

import (
	"fmt"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/energomonitor/bisquitt"
)

const (
	HostFlag                    = "host"
	PortFlag                    = "port"
	DtlsFlag                    = "dtls"
	SelfSignedFlag              = "self-signed"
	PskFlag                     = "psk"
	PskCacheExpirationFlag      = "psk-cache-expiration"
	PskIdentityFlag             = "psk-identity"
	PSKAPITimeoutFlag           = "psk-api-timeout"
	PSKAPIBasicAuthUsernameFlag = "psk-api-basic-auth-username"
	PSKAPIBasicAuthPasswordFlag = "psk-api-basic-auth-password"
	PSKAPIEndpointFlag          = "psk-api-endpoint"
	CertFlag                    = "cert"
	KeyFlag                     = "key"
	CAFileFlag                  = "cafile"
	CAPathFlag                  = "capath"
	InsecureFlag                = "insecure"
	DebugFlag                   = "debug"
	TopicFlag                   = "topic"
	PredefinedTopicFlag         = "predefined-topic"
	PredefinedTopicsFileFlag    = "predefined-topics-file"
	QOSFlag                     = "qos"
	ClientIDFlag                = "client-id"
	WillTopicFlag               = "will-topic"
	WillMessageFlag             = "will-message"
	WillQOSFlag                 = "will-qos"
	WillRetainFlag              = "will-retain"
	UserFlag                    = "user"
	PasswordFlag                = "password"
)

func init() {
	cli.HelpFlag = &cli.BoolFlag{
		Name:  "help",
		Usage: "show this help",
	}
}

var Application = cli.App{
	Name:        "bisquitt-sub",
	Usage:       "A MQTT-SN client with DTLS support",
	ArgsUsage:   " ",
	Version:     bisquitt.Version(),
	Description: "A MQTT-SN client with DTLS support.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    HostFlag,
			Aliases: []string{"h"},
			Usage:   "MQTT-SN broker host",
			Value:   "127.0.0.1",
			EnvVars: []string{
				"HOST",
			},
		},
		&cli.IntFlag{
			Name:    PortFlag,
			Aliases: []string{"p"},
			Usage:   "MQTT-SN broker port",
			Value:   1883,
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
			Name:  PskFlag,
			Usage: "use PSK",
			EnvVars: []string{
				"PSK_ENABLED",
			},
		},
		&cli.DurationFlag{
			Name:  PskCacheExpirationFlag,
			Value: 5 * time.Minute,
			Usage: "PSKKeys cache expiration",
			EnvVars: []string{
				"PSK_CACHE_EXPIRATION",
			},
		},
		&cli.StringFlag{
			Name:  PskIdentityFlag,
			Usage: "PSKKeys identity",
			EnvVars: []string{
				"PSK_IDENTITY",
			},
		},
		&cli.DurationFlag{
			Name:  PSKAPITimeoutFlag,
			Usage: "PSKKeys API timeout",
			Value: 5 * time.Second,
			EnvVars: []string{
				"PSK_API_TIMEOUT",
			},
		},
		&cli.StringFlag{
			Name:  PSKAPIBasicAuthUsernameFlag,
			Usage: "PSKKeys API basic auth username",
			EnvVars: []string{
				"PSK_API_BASIC_AUTH_USERNAME",
			},
		},
		&cli.StringFlag{
			Name:  PSKAPIBasicAuthPasswordFlag,
			Usage: "PSKKeys API basic auth password",
			EnvVars: []string{
				"PSK_API_BASIC_AUTH_PASSWORD",
			},
		},
		&cli.StringFlag{
			Name:  PSKAPIEndpointFlag,
			Usage: "PSKKeys API endpoint",
			EnvVars: []string{
				"PSK_API_ENDPOINT",
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
		&cli.PathFlag{
			Name:  CAFileFlag,
			Usage: "CA certificate file",
			EnvVars: []string{
				"CA_FILE",
			},
		},
		&cli.PathFlag{
			Name:  CAPathFlag,
			Usage: "CA certificates directory",
			EnvVars: []string{
				"CA_PATH",
			},
		},
		&cli.BoolFlag{
			Name:  InsecureFlag,
			Usage: "do not check server certificate",
			EnvVars: []string{
				"INSECURE",
			},
		},
		&cli.BoolFlag{
			Name:    DebugFlag,
			Aliases: []string{"d"},
			Usage:   "print debug messages",
			EnvVars: []string{
				"DEBUG",
			},
		},
		&cli.StringSliceFlag{
			Name:    TopicFlag,
			Aliases: []string{"t"},
			Usage:   "topic(s) to subscribe to",
			EnvVars: []string{
				"TOPICS",
			},
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:  PredefinedTopicFlag,
			Usage: fmt.Sprintf("predefined topic, takes precedence over --%s (format: topicName;topicID)", PredefinedTopicsFileFlag),
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
		&cli.UintFlag{
			Name:    QOSFlag,
			Aliases: []string{"q"},
			Usage:   "quality of service (0-2)",
			Value:   0,
			EnvVars: []string{
				"QOS",
			},
		},
		&cli.StringFlag{
			Name:    ClientIDFlag,
			Aliases: []string{"i"},
			Usage:   `client ID to use, defaults to randomized string`,
			Value:   "",
			EnvVars: []string{
				"CLIENT_ID",
			},
		},
		&cli.StringFlag{
			Name:  WillTopicFlag,
			Usage: "topic to send the will message to",
			Value: "",
			EnvVars: []string{
				"WILL_TOPIC",
			},
		},
		&cli.StringFlag{
			Name:  WillMessageFlag,
			Usage: "will message",
			Value: "",
			EnvVars: []string{
				"WILL_MESSAGE",
			},
		},
		&cli.UintFlag{
			Name:  WillQOSFlag,
			Usage: "will message quality of service (0-3) Use 3 for QoS -1",
			Value: 0,
			EnvVars: []string{
				"WILL_QOS",
			},
		},
		&cli.BoolFlag{
			Name:  WillRetainFlag,
			Usage: "will message should be retained",
			EnvVars: []string{
				"WILL_RETAINED",
			},
		},
		&cli.StringFlag{
			Name:    UserFlag,
			Aliases: []string{"u"},
			Usage:   `username to use for non-standard AUTH authentication`,
			EnvVars: []string{
				// Don't use "USER" here because USER env var contains current user.
				"USERNAME",
			},
		},
		&cli.StringFlag{
			Name:    PasswordFlag,
			Aliases: []string{"P"},
			Usage:   `password to use for non-standard AUTH authentication`,
			EnvVars: []string{
				"PASSWORD",
			},
		},
	},
	UseShortOptionHandling: true,
	HideHelpCommand:        true,
	Action:                 handleAction(),
}
