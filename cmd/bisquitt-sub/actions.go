package main

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/urfave/cli/v2"

	snClient "github.com/energomonitor/bisquitt/client"
	pkts1 "github.com/energomonitor/bisquitt/packets1"
	"github.com/energomonitor/bisquitt/topics"
	"github.com/energomonitor/bisquitt/util"
	cryptoutils "github.com/energomonitor/bisquitt/util/crypto"
)

func handleAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		if c.Uint(QOSFlag) > 2 {
			return fmt.Errorf("QOS must be 0-2, got %v", c.Uint(QOSFlag))
		}
		qos := uint8(c.Uint(QOSFlag))
		topicList := c.StringSlice(TopicFlag)

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
		caFile := c.Path(CAFileFlag)
		caPath := c.Path(CAPathFlag)
		debug := c.Bool(DebugFlag)

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

		var caFiles []string
		if caFile != "" {
			caFiles = append(caFiles, caFile)
		}

		if caPath != "" {
			globPattern := path.Join(caPath, "*.crt")
			files, err := filepath.Glob(globPattern)
			if err != nil {
				return fmt.Errorf("loading CA certificates failed: glob error: %s", err)
			}
			caFiles = append(caFiles, files...)
		}

		var caCertificates []*x509.Certificate
		for _, file := range caFiles {
			certs, err := cryptoutils.LoadX509Certificate(file)
			if err != nil {
				return fmt.Errorf("parsing a CA certificate '%s' failed: %s", file, err)
			}
			caCertificates = append(caCertificates, certs...)
		}

		predefinedTopics := topics.PredefinedTopics{}

		if c.IsSet(PredefinedTopicsFileFlag) {
			v, err := topics.ParsePredefinedTopicOptions(c.Path(PredefinedTopicsFileFlag))
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

		brokerHost := c.String(HostFlag)
		brokerPort := c.Int(PortFlag)
		if useDTLS && !c.IsSet(PortFlag) {
			brokerPort = 8883
		}
		brokerAddress := fmt.Sprintf("%s:%d", brokerHost, brokerPort)

		insecure := c.Bool(InsecureFlag)

		var clientID string
		if c.IsSet(ClientIDFlag) {
			clientID = c.String(ClientIDFlag)
		} else {
			clientID = fmt.Sprintf("bisquitt-sub-%x", rand.Uint64())
		}

		var user string
		if c.IsSet(UserFlag) {
			user = c.String(UserFlag)
			if user == "" {
				return fmt.Errorf(`"--%s" must not be empty`, UserFlag)
			}
			if !useDTLS && !insecure {
				return fmt.Errorf(
					`Using plain text auth without DTLS is insecure. Use "%s" to use anyway.`,
					InsecureFlag,
				)
			}
		}
		password := []byte(c.String(PasswordFlag))

		clientCfg := &snClient.ClientConfig{
			ClientID:                clientID,
			UseDTLS:                 useDTLS,
			SelfSigned:              useSelfSigned,
			UsePSK:                  usePSK,
			PSKKeys:                 cache.New(pskCacheExpiration, 5*time.Minute),
			PSKCacheExpiration:      pskCacheExpiration,
			PSKIdentityHint:         pskIdentity,
			PSKAPITimeout:           pskAPITimeout,
			PSKAPIBasicAuthUsername: pskAPIBasicAuthUsername,
			PSKAPIBasicAuthPassword: pskAPIBasicAuthPassword,
			PSKAPIEndpoint:          pskAPIEndpoint,
			Insecure:                insecure,
			Certificate:             certificate,
			PrivateKey:              privateKey,
			CACertificates:          caCertificates,
			PredefinedTopics:        predefinedTopics,
			RetryDelay:              10 * time.Second,
			RetryCount:              4,
			ConnectTimeout:          20 * time.Second,
			KeepAlive:               60 * time.Second,
			CleanSession:            true,
			User:                    user,
			Password:                password,
		}

		if c.IsSet(WillTopicFlag) {
			willTopic := c.String(WillTopicFlag)
			if willTopic == "" {
				return errors.New("will topic must not be empty if set")
			}
			clientCfg.WillTopic = willTopic
			clientCfg.WillPayload = []byte(c.String(WillMessageFlag))

			qos := c.Uint(WillQOSFlag)
			if qos > 2 {
				return fmt.Errorf("will QOS must be 0-2, got %d", qos)
			}
			clientCfg.WillQOS = uint8(qos)
			clientCfg.WillRetained = c.Bool(WillRetainFlag)
		}

		var logger util.Logger
		if debug {
			logger = util.NewDebugLogger("sub")
		} else {
			logger = util.NewProductionLogger("sub")
		}
		defer logger.Sync()

		client := snClient.NewClient(logger, clientCfg)

		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

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
			client.Close()
		}()

		if err := client.Dial(brokerAddress); err != nil {
			return fmt.Errorf("cannot connect to MQTT-SN broker: %s", err)
		}
		defer client.Close()

		if err := client.Connect(); err != nil {
			return err
		}

		handler := func(client *snClient.Client, topic string, msg *pkts1.Publish) {
			var flags []string
			if msg.Retain {
				flags = append(flags, "retained")
			}
			if msg.QOS != 0 {
				flags = append(flags, fmt.Sprintf("qos %d", msg.QOS))
			}
			var flagsStr string
			if len(flags) > 0 {
				flagsStr = "[" + strings.Join(flags, ", ") + "]"
			}
			fmt.Printf("%s: %s %s\n", topic, msg.Data, flagsStr)
		}

		for _, topic := range topicList {
			topicID, isPredefinedTopic := predefinedTopics.GetTopicID(clientID, topic)
			if isPredefinedTopic {
				if err := client.SubscribePredefined(topicID, qos, handler); err != nil {
					return err
				}
			} else {
				if err := client.Subscribe(topic, qos, handler); err != nil {
					return err
				}
			}
		}

		// Wait until client quits.
		if err := client.Wait(); err != nil {
			return err
		}

		return nil
	}
}
