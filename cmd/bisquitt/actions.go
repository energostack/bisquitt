package main

import (
	"context"
	"crypto"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/urfave/cli/v2"

	"github.com/energomonitor/bisquitt/gateway"
	"github.com/energomonitor/bisquitt/topics"
	"github.com/energomonitor/bisquitt/util"
	cryptoutils "github.com/energomonitor/bisquitt/util/crypto"
	"github.com/energomonitor/bisquitt/util/platform"
)

func handleAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		useDTLS := c.Bool(DtlsFlag)
		useSelfSigned := c.Bool(SelfSignedFlag)
		usePSK := c.Bool(PskFlag)
		pskCacheExpiration := c.Duration(PskCacheExpirationFlag)
		pskIdentity := c.String(PskIdentityFlag)
		pskAPITimeout := c.Duration(PSKAPITimeoutFlag)
		pskAPIBasicAuthUsername := c.String(PSKAPIBasicAuthUsernameFlag)
		pskAPIBasicAuthPassword := c.String(PSKAPIBasicAuthPasswordFlag)
		pskAPIEndpoint := c.String(PSKAPIEndpointFlag)
		certFile := c.Path(CertFlag)
		keyFile := c.Path(KeyFlag)
		debug := c.Bool(DebugFlag)
		syslog := c.Bool(SyslogFlag)

		if useDTLS && ((certFile == "" || keyFile == "") && !useSelfSigned) && !usePSK {
			return fmt.Errorf(`options "--%s" and "--%s" are mandatory when using DTLS. Use "--%s" to generate self-signed certificate.`,
				CertFlag, KeyFlag, SelfSignedFlag)
		}

		var certificate *tls.Certificate
		var privateKey crypto.PrivateKey
		if certFile != "" && keyFile != "" {
			cert, err := cryptoutils.LoadCertificate(certFile)
			if err != nil {
				return fmt.Errorf("cannot load a certificate from file '%s': %s", certFile, err)
			}
			certificate = cert

			key, err := cryptoutils.LoadKey(keyFile)
			if err != nil {
				return fmt.Errorf("cannot load a private key from file '%s': %s", keyFile, err)
			}
			privateKey = key
		}

		authEnabled := c.Bool(AuthFlag)
		if authEnabled && !useDTLS && !c.Bool(InsecureFlag) {
			return fmt.Errorf(`Using plain text auth without DTLS is insecure. Use "--%s" to use anyway.`, InsecureFlag)
		}

		predefinedTopics := topics.PredefinedTopics{}

		if c.IsSet(PredefinedTopicsFileFlag) {
			v, err := topics.ReadPredefinedTopicsFile(c.Path(PredefinedTopicsFileFlag))
			if err != nil {
				return err
			}
			predefinedTopics = v
		}
		if c.IsSet(PredefinedTopicFlag) {
			v, err := topics.ParsePredefinedTopicOptions(c.StringSlice(PredefinedTopicFlag)...)
			if err != nil {
				return fmt.Errorf(`parsing "--%s" failed: %s`, PredefinedTopicFlag, err)
			}
			predefinedTopics.Merge(v)
		}

		host := c.String(HostFlag)
		port := c.Int(PortFlag)
		if useDTLS && !c.IsSet(PortFlag) {
			port = 8883
		}

		mqttBrokerHost := c.String(MqttHostFlag)
		mqttBrokerPort := c.Int(MqttPortFlag)

		// In MQTT, a username and password can be set or unset. At least the
		// password can also be empty:
		//
		// The Password field contains 0 to 65535 bytes of binary data
		// [MQTT 3.1.1 specification, chapter 3.1.3.5 Password]
		//
		// We are using the `*string` default value `nil` as a value for "unset".
		var mqttUser *string
		if c.IsSet(MqttUserFlag) {
			mqttUser2 := c.String(MqttUserFlag)
			mqttUser = &mqttUser2
		}
		var mqttPassword []byte
		if c.IsSet(MqttPasswordFlag) {
			mqttPassword = []byte(c.String(MqttPasswordFlag))
		}

		if c.IsSet(MqttPasswordFileFlag) {
			var err error
			mqttPassword, err = ioutil.ReadFile(c.Path(MqttPasswordFileFlag))
			if err != nil {
				return fmt.Errorf("cannot read password file: %s", err)
			}
		}

		mqttBrokerAddress, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", mqttBrokerHost, mqttBrokerPort))
		if err != nil {
			return err
		}
		mqttConnectionTimeout := c.Duration(MqttTimeoutFlag)

		performanceLogTime := c.Duration(PerformanceLogTimeFlag)

		gwConfig := &gateway.GatewayConfig{
			MqttBrokerAddress:       mqttBrokerAddress,
			MqttConnectionTimeout:   mqttConnectionTimeout,
			MqttUser:                mqttUser,
			MqttPassword:            mqttPassword,
			UseDTLS:                 useDTLS,
			UsePSK:                  usePSK,
			PSKKeys:                 cache.New(pskCacheExpiration, 5*time.Minute),
			PSKCacheExpiration:      pskCacheExpiration,
			PSKIdentityHint:         pskIdentity,
			PSKAPITimeout:           pskAPITimeout,
			PSKAPIBasicAuthUsername: pskAPIBasicAuthUsername,
			PSKAPIBasicAuthPassword: pskAPIBasicAuthPassword,
			PSKAPIEndpoint:          pskAPIEndpoint,
			SelfSigned:              useSelfSigned,
			Certificate:             certificate,
			PrivateKey:              privateKey,
			PerformanceLogTime:      performanceLogTime,
			PredefinedTopics:        predefinedTopics,
			AuthEnabled:             authEnabled,
			RetryDelay:              10 * time.Second,
			RetryCount:              4,
		}

		logTag := "gw"
		var logger util.Logger
		if syslog {
			logger, err = util.NewSyslogLogger(logTag, debug)
			if err != nil {
				return fmt.Errorf("cannot initialize syslog: %s", err)
			}
		} else {
			if debug {
				logger = util.NewDebugLogger(logTag)
			} else {
				logger = util.NewProductionLogger(logTag)
			}
		}
		defer logger.Sync()

		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			s := "signal"
			switch <-signalCh {
			case syscall.SIGTERM:
				s = "SIGTERM"
			case syscall.SIGHUP:
				s = "SIGHUP"
			case syscall.SIGINT:
				s = "SIGINT"
			}
			logger.Info("%s caught", s)
			cancel()
		}()

		logger.Info("%s version %s starting", c.App.Name, c.App.Version)

		if c.IsSet(GroupFlag) || c.IsSet(UserFlag) {
			if c.IsSet(GroupFlag) {
				group := c.String(GroupFlag)
				if err := platform.SetGroup(group); err != nil {
					return fmt.Errorf("failed to setgid: %s", err)
				}
			}
			if c.IsSet(UserFlag) {
				user := c.String(UserFlag)
				if err := platform.SetUser(user); err != nil {
					return fmt.Errorf("failed to setuid: %s", err)
				}
			}

			currentUser, err := platform.GetCurrentUser()
			if err != nil {
				return fmt.Errorf("failed to get the current user name: %s", err)
			}
			currentGroup, err := platform.GetCurrentGroup()
			if err != nil {
				return fmt.Errorf("failed to get the current group name: %s", err)
			}

			logger.Info("switched to %s:%s", currentUser.Username, currentGroup.Name)
		}

		gw := gateway.NewGateway(logger, gwConfig)

		return gw.ListenAndServe(ctx, fmt.Sprintf("%s:%d", host, port))
	}
}
