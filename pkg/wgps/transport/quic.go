package transport

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"

	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/utils"
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
		//fmt.Println("listening")
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
		fmt.Println("Opened Stream")
		if err != nil {

			return err
		}
		//fmt.Println("Initiated stream")
		var initBytes []byte
		initBytes = append(initBytes, utils.BigIntToBytes(uint64(len(string(i))))...) // Length of the string
		initBytes = append(initBytes, byte(i))                                        // The string
		_, err = q.InitiatedStreams[i].Write(initBytes)
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

func (q *QuicTransport) IsClosed() bool {
	return q.Closed
}

func (q QuicTransport) Recv(writeTo chan []byte, channel wgpstypes.Channel, role wgpstypes.SyncRole) {
	dataChan := make(chan []byte)
	go func() {
		if wgpstypes.IsAlfie(role) {
			for {
				buffer := make([]byte, 8) // Adjust buffer size as needed
				if q.InitiatedStreams[channel] != nil {
					_, err := io.ReadFull(q.InitiatedStreams[channel], buffer)
					if err != nil {
						// Handle Error
						return
					}
					fmt.Println(buffer)
					bytesWithLen := binary.BigEndian.Uint64(buffer[:8])
					fmt.Printf("Length of message is %v\n", bytesWithLen)
					newBuffer := make([]byte, bytesWithLen)
					if q.InitiatedStreams[channel] != nil {
						n, err := io.ReadFull(q.InitiatedStreams[channel], newBuffer)
						if err != nil {
							// Handle Error
							return
						}

						fmt.Printf("Writing message now %v\n", string(newBuffer[:n]))
						writeTo <- newBuffer[:n]
					}
				}
			}
		} else {
			for {
				buffer := make([]byte, 8)
				if q.AcceptedStreams[channel] != nil {
					_, err := io.ReadFull(q.AcceptedStreams[channel], buffer)
					if err != nil {
						// Handle Error
						return
					}
					fmt.Println(buffer)
					fmt.Printf("Length is %v\n", binary.BigEndian.Uint64(buffer))
					bytesWithLen := binary.BigEndian.Uint64(buffer[:8])
					newBuffer := make([]byte, bytesWithLen)
					if q.AcceptedStreams[channel] != nil {
						n, err := io.ReadFull(q.AcceptedStreams[channel], newBuffer)
						if err != nil {
							// Handle Error
							return
						}

						fmt.Printf("Writing message now %v\n", string(newBuffer[:n]))
						writeTo <- newBuffer[:n]
					}
				}
			}
		}
	}()
	for {
		for data := range dataChan {
			writeTo <- data
		}
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
