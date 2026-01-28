package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/jackc/pgx/v5/pgproto3"
	"log"
	"math/big"
	"net"
	"time"
)

func GenerateSelfSignedCert() (*tls.Config, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("failed to encode private key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to create X509 key pair: %w", err)
	}

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
	return tlsConfig, nil
}

func handleSSLRequest(conn net.Conn) (*tls.Conn, error) {
	_, err := conn.Write([]byte("S"))
	if err != nil {
		return nil, fmt.Errorf("failed to write SSLResponse: %w", err)
	}

	// Generate a self-signed cert
	tlsConfig, err := GenerateSelfSignedCert()
	if err != nil {
		return nil, fmt.Errorf("failed to generate self-signed certificate: %w", err)
	}

	tlsConn := tls.Server(conn, tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		return nil, fmt.Errorf("failed TLS handshake: %w", err)
	}
	return tlsConn, nil
}

func upgradeBackendToTls(backend pgproto3.Backend, client net.Conn) (*pgproto3.Backend, error) {
	msg, err := backend.ReceiveStartupMessage()
	if err != nil {
		log.Printf("failed to receive startup message: %v", err)
		return nil, err
	}
	switch msg.(type) {
	case *pgproto3.SSLRequest:
		tlsConn, err := handleSSLRequest(client)
		if err != nil {
			return nil, fmt.Errorf("failed to handle SSL request: %w", err)
		}
		newBackend := pgproto3.NewBackend(tlsConn, tlsConn)
		return newBackend, nil
	default:
		fmt.Printf("Received unexpected message: %T\n", msg)
		return nil, fmt.Errorf("unsupported message type: %T", msg)
	}
}
