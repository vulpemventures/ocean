package grpc_interface

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"net"
	"path/filepath"

	"golang.org/x/net/http2"
)

const (
	minPort = 1024
	maxPort = 49151
)

type ServiceConfig struct {
	Port         int
	NoTLS        bool
	TLSLocation  string
	ExtraIPs     []string
	ExtraDomains []string
}

func (c ServiceConfig) validate() error {
	if c.Port < minPort || c.Port > maxPort {
		return fmt.Errorf("port must be in range [%d, %d]", minPort, maxPort)
	}
	return nil
}

func (c ServiceConfig) insecure() bool {
	return c.NoTLS
}

func (c ServiceConfig) address() string {
	return fmt.Sprintf(":%d", c.Port)
}

func (c ServiceConfig) listener() net.Listener {
	lis, _ := net.Listen("tcp", c.address())

	if c.insecure() {
		return lis
	}
	return tls.NewListener(lis, c.tlsConfig())
}

func (c ServiceConfig) tlsConfig() *tls.Config {
	if c.insecure() {
		return nil
	}
	cert, _ := tls.LoadX509KeyPair(c.tlsCertPath(), c.tlsKeyPath())
	return &tls.Config{
		NextProtos:   []string{"http/1.1", http2.NextProtoTLS, "h2-14"}, // h2-14 is just for compatibility. will be eventually removed.
		Certificates: []tls.Certificate{cert},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		Rand: rand.Reader,
	}
}

func (c ServiceConfig) tlsKeyPath() string {
	return filepath.Join(c.TLSLocation, tlsKeyFile)
}

func (c ServiceConfig) tlsCertPath() string {
	return filepath.Join(c.TLSLocation, tlsCertFile)
}
