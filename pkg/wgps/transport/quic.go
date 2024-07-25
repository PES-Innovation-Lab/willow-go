package transport

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"

	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/quic-go/quic-go"
)

//const addr = "0.0.0.0:4242"

type QuicTransport struct {
	AcceptedStreams  []quic.Stream
	InitiatedStreams []quic.Stream

	Closed bool
}

func NewQuicTransport(addr string) (*QuicTransport, error) {
	acceptedStreams := make([]quic.Stream, 8)
	initatedStreams := make([]quic.Stream, 8)
	newQuicTransport := &QuicTransport{
		AcceptedStreams:  acceptedStreams,
		InitiatedStreams: initatedStreams,
		Closed:           false,
	}

	go func() {
		listener, err := quic.ListenAddr(addr, generateTLSConfig(), nil)
		if err != nil {
			log.Fatalf("Failed to set up listener: %v", err)
		}
		fmt.Println("listening")
		var conn quic.Connection
		conn, err = listener.Accept(context.Background())

		if err != nil {
			log.Fatalf("Failed to set up connection: %v", err)
		}

		for i := 0; i < 8; i++ {
			newQuicTransport.AcceptedStreams[i], err = conn.AcceptStream(context.Background())
			fmt.Println("Accepted stream")
			if err != nil {
				log.Fatalf("Failed to set up stream: %v", err)
			}
		}

	}()

	return newQuicTransport, nil

}

func (q *QuicTransport) Initiate(addr string) error {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"Willow-Go-Quic"},
	}
	var conn quic.Connection
	conn, err := quic.DialAddr(context.Background(), addr, tlsConf, nil)
	if err != nil {
		return err
	}

	for i := 0; i < 8; i++ {
		q.InitiatedStreams[i], err = conn.OpenStreamSync(context.Background())
		if err != nil {

			return err
		}
		fmt.Println("Initiated stream")
		_, err = q.InitiatedStreams[i].Write([]byte{byte(i)})
		if err != nil {
			return err
		}
	}
	return nil
}

func (q *QuicTransport) Send(data []byte, channel wgpstypes.Channel, role wgpstypes.SyncRole) error {
	if q.Closed {
		return fmt.Errorf("transport is closed")
	}
	if wgpstypes.IsAlfie(role) {
		_, err := q.InitiatedStreams[channel].Write(data)
		if err != nil {
			return err
		}
	} else {
		_, err := q.AcceptedStreams[channel].Write(data)
		if err != nil {
			return err
		}
	}

	return nil

}

func (t *QuicTransport) Close() error {

	// Close each stream
	for _, stream := range t.AcceptedStreams {
		if err := stream.Close(); err != nil {
			return err
		}
	}
	for _, stream := range t.InitiatedStreams {
		if err := stream.Close(); err != nil {
			return err
		}
	}
	// Mark the transport as closed
	t.Closed = true

	return nil
}

func (q QuicTransport) IsClosed() bool {
	return q.Closed
}

func (q *QuicTransport) Recv(channel wgpstypes.Channel, role wgpstypes.SyncRole) ([]byte, error) {
	buffer := make([]byte, 4096) // Adjust buffer size as needed
	if wgpstypes.IsAlfie(role) {
		n, err := q.InitiatedStreams[channel].Read(buffer)
		if err != nil {
			return nil, err
		}
		return buffer[:n], nil
	} else {
		n, err := q.AcceptedStreams[channel].Read(buffer)
		if err != nil {
			return nil, err
		}
		return buffer[:n], nil
	}

}

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"Willow-Go-Quic"},
	}
}
