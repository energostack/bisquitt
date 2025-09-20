// Package gateway implements a MQTT-SN version 1.2 transparent gateway with
// optional DTLS encryption.
package gateway

import (
	"context"
	"crypto"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pion/dtls/v2"
	"github.com/pion/dtls/v2/pkg/crypto/selfsign"
	"github.com/pion/udp"

	"github.com/energostack/bisquitt/topics"
	"github.com/energostack/bisquitt/util"
)

type GatewayConfig struct {
	MqttBrokerAddress     *net.TCPAddr
	MqttConnectionTimeout time.Duration
	MqttUser              *string
	MqttPassword          []byte
	// UsePSK controls whether pre-shared key should be used to secure the
	// connection to the MQTT-SN gateway. If UsePSK is true, you must provide
	// PSKIdentityHint, PSKAPIBasicAuthUsername, PSKAPIBasicAuthPassword and
	// PSKAPIEndpoint.
	// If UsePSK is true, the client will use PSKKeys instead of the certificate
	// and private key.
	UsePSK                  bool
	PSKKeys                 *cache.Cache
	PSKCacheExpiration      time.Duration
	PSKIdentityHint         string
	PSKAPITimeout           time.Duration
	PSKAPIBasicAuthUsername string
	PSKAPIBasicAuthPassword string
	PSKAPIEndpoint          string
	UseDTLS                 bool
	SelfSigned              bool
	Certificate             *tls.Certificate
	PrivateKey              crypto.PrivateKey
	PerformanceLogTime      time.Duration
	PredefinedTopics        topics.PredefinedTopics
	AuthEnabled             bool
	// TRetry in MQTT-SN specification
	RetryDelay time.Duration
	// NRetry in MQTT-SN specification
	RetryCount uint
}

type Gateway struct {
	cfg *GatewayConfig
	log util.Logger
}

// Timeout for DTLS connection establishment.
const dtlsConnectTimeout = 300 * time.Second

func NewGateway(log util.Logger, cfg *GatewayConfig) *Gateway {
	return &Gateway{
		cfg: cfg,
		log: log,
	}
}

func newDTLSListener(ctx context.Context, cfg *GatewayConfig, address *net.UDPAddr) (net.Listener, error) {
	var certificate *tls.Certificate
	var err error

	logger := util.NewProductionLogger("gateway")

	if !cfg.UsePSK && cfg.UseDTLS {
		if cfg.SelfSigned {
			var cert tls.Certificate
			cert, err = selfsign.GenerateSelfSigned()
			certificate = &cert
		} else {
			privateKey := cfg.PrivateKey
			if privateKey == nil {
				err = errors.New("private key is missing")
			}

			certificate = cfg.Certificate
			if certificate == nil {
				err = errors.New("TLS certificate is missing")
			} else {
				certificate.PrivateKey = privateKey
			}
		}
	}

	if err != nil {
		return nil, err
	}

	dtlsConfig := &dtls.Config{
		ExtendedMasterSecret: dtls.RequireExtendedMasterSecret,
		ConnectContextMaker: func() (context.Context, func()) {
			return context.WithTimeout(ctx, dtlsConnectTimeout)
		},
	}

	if !cfg.UsePSK && cfg.UseDTLS && certificate != nil {
		dtlsConfig.Certificates = []tls.Certificate{*certificate}
	}

	if cfg.UsePSK && cfg.UseDTLS {
		dtlsConfig.CipherSuites = []dtls.CipherSuiteID{dtls.TLS_PSK_WITH_AES_128_GCM_SHA256}
		dtlsConfig.PSK = func(hint []byte) ([]byte, error) {
			psk, ok := cfg.PSKKeys.Get(string(hint))
			if ok {
				return psk.([]byte), nil
			}

			psk, ok = util.GetPSKKeyFromAPI(
				string(hint),
				cfg.PSKAPIEndpoint,
				cfg.PSKAPIBasicAuthUsername,
				cfg.PSKAPIBasicAuthPassword,
				cfg.PSKAPITimeout,
				logger,
			)

			if ok {
				cfg.PSKKeys.Set(string(hint), psk, cfg.PSKCacheExpiration)
				return psk.([]byte), nil
			}

			return nil, errors.New("PSK key not found")
		}
		dtlsConfig.PSKIdentityHint = []byte(cfg.PSKIdentityHint)
	}

	return dtls.Listen("udp", address, dtlsConfig)
}

func newUDPListener(ctx context.Context, address *net.UDPAddr) (net.Listener, error) {
	udpConfig := &udp.ListenConfig{}
	return udpConfig.Listen("udp", address)
}

// ListenAndServe starts a gateway listening on the given address. It returns
// only on fatal internal errors or when the given context is canceled.
func (gw *Gateway) ListenAndServe(ctx context.Context, address string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}

	var snListener net.Listener
	if gw.cfg.UseDTLS {
		snListener, err = newDTLSListener(ctx, gw.cfg, udpAddr)
	} else {
		snListener, err = newUDPListener(ctx, udpAddr)
	}
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		snListener.Close()
	}()

	gw.log.Info("Listening on %s", snListener.Addr().String())

	handlerCfg := &handlerConfig{
		MqttBrokerAddress:     gw.cfg.MqttBrokerAddress,
		MqttUser:              gw.cfg.MqttUser,
		MqttPassword:          gw.cfg.MqttPassword,
		MqttConnectionTimeout: gw.cfg.MqttConnectionTimeout,
		AuthEnabled:           gw.cfg.AuthEnabled,
		RetryDelay:            gw.cfg.RetryDelay,
		RetryCount:            gw.cfg.RetryCount,
	}

	for {
		clientConn, err := snListener.Accept()
		if err != nil {
			if _, ok := err.(*dtls.HandshakeError); ok {
				gw.log.Error("Client TLS handshake error: %s", err)
				continue
			}
			if err == udp.ErrClosedListener {
				return nil
			}
			gw.log.Error("MQTT-SN Accept error: %v", err)
			return err
		}
		gw.log.Debug("Client connected: %s", clientConn.RemoteAddr().String())
		handlerID := clientConn.RemoteAddr().String()
		handlerLogger := gw.log.WithTag(fmt.Sprintf("h:%s", handlerID))
		handler := newHandler(handlerCfg, gw.cfg.PredefinedTopics, handlerLogger)
		go func() {
			defer func() {
				handlerLogger.Debug("Closing MQTT-SN connection")
				err := clientConn.Close()
				if err != nil {
					handlerLogger.Error("Error closing MQTT-SN connection: %s", err)
				}
			}()

			handler.run(ctx, clientConn)
		}()
	}
}
