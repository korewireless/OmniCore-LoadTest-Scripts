package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"sync/atomic"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var messageId uint64

type PayloadGenerator func(i int) string

func GenerateMessageBaseValue() {
	rand.Seed(time.Now().Unix())
	messageId = randomSource.Uint64()
}

func defaultPayloadGen() PayloadGenerator {
	return func(i int) string {
		return fmt.Sprintf("this is msg #%d!", i)
	}
}

func constantPayloadGenerator(payload string) PayloadGenerator {
	return func(i int) string {
		return payload
	}
}

func filePayloadGenerator(filepath string) PayloadGenerator {
	inputPath := strings.Replace(filepath, "@", "", 1)
	content, err := ioutil.ReadFile(inputPath)
	if err != nil {
		fmt.Printf("error reading payload file: %v\n", err)
		os.Exit(1)
	}
	return func(i int) string {
		return string(content)
	}
}

type Worker struct {
	WorkerId             int
	BrokerUrl            string
	Username             string
	Password             string
	SkipTLSVerification  bool
	NumberOfMessages     int
	PayloadGenerator     PayloadGenerator
	Timeout              time.Duration
	Retained             bool
	PublisherQoS         byte
	SubscriberQoS        byte
	CA                   []byte
	Cert                 []byte
	Key                  []byte
	PauseBetweenMessages time.Duration
}

func setSkipTLS(o *mqtt.ClientOptions) {
	oldTLSCfg := o.TLSConfig
	if oldTLSCfg == nil {
		oldTLSCfg = &tls.Config{}
	}
	oldTLSCfg.InsecureSkipVerify = true
	o.SetTLSConfig(oldTLSCfg)
}
func NewTlsConfig2() *tls.Config {
	certpool := x509.NewCertPool()
	ca := "-----BEGIN CERTIFICATE-----\nMIIDVjCCAj6gAwIBAgITZSbEZ/rEg6U4OYgwH8okbPq1OTANBgkqhkiG9w0BAQsF\nADAyMRUwEwYDVQQKEwxLb3Jld2lyZWxlc3MxGTAXBgNVBAMTEGtvcmV3aXJlbGVz\ncy5jb20wIBcNMjIxMTExMDY0ODA2WhgPMjA1MjExMDMwNjQ4MDVaMDIxFTATBgNV\nBAoTDEtvcmV3aXJlbGVzczEZMBcGA1UEAxMQa29yZXdpcmVsZXNzLmNvbTCCASIw\nDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAIs2ZN6edS0d+UkhQbSJrgC1ZwBi\n+XrumuECOvz9Giwr2nRbsZofjGYAy+PdKfkzlJD6aOWKtX9tx5NWlihbSkGbc8kj\nYQlLwrm/6/gZWHjZeMY8rj18+ieIzFB/Y3sK3LmkrGnmpw1FQJtTDoOf0S6YfWIX\ngRv1qJrZevW7nrzlzzex3dmJroj9jAcyxgj+VK7IrBfUJTq4vQ4w6ltPHKh9ZxNL\nyIaHb94BUpPugXecwAyZuKjEFPH8z62bwo9uSnJwogshFueIAF9Nw57J+UgsCFrY\nqJHLz4DSYlhZhAaqoYDCDUIaz5xQkW0ggywqOYrkB7RK+r1W6c8MIgF4JI8CAwEA\nAaNjMGEwDgYDVR0PAQH/BAQDAgEGMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYE\nFBm2Uovs4uqrVGW7OkUmnPXv3BBFMB8GA1UdIwQYMBaAFBm2Uovs4uqrVGW7OkUm\nnPXv3BBFMA0GCSqGSIb3DQEBCwUAA4IBAQBFBeKwG6l528TjrGjNlGumkBjFaPK0\n5GS5/MWBEqOzXRzA1zs2XpEl4/HH7LdOtJvI2S+oSkbOYOVF3hJoh7H/Z/jOsbDk\ns/OXu+BDPnJN6RgmtcqDAZa/KtKLhcIGCwK5Sl/C+Vx6bUogXTCai8RwnGB9XwF1\n6EUZcyaaZpcJg0wEjYUm/tyvehnpR8Usl9aDUdpKM0NgTgCdWaAQCNT4HpTvHe1v\n8R97c9OpcTBc/Cs2XadBtOY2lwDHskMij50+n+xzas+jBPbrnTbYIjcixSEiSsxm\n5vVHrDWO9naPgbO7sIuExvhJSBMeBeQBQ4wUFma5Endcg4ZE8CCqGlsn\n-----END CERTIFICATE-----\n"
	bytesCa := []byte(ca)
	certpool.AppendCertsFromPEM(bytesCa)
	return &tls.Config{
		RootCAs: certpool,
	}
}
func NewTLSConfig(ca, certificate, privkey []byte) (*tls.Config, error) {
	// Import trusted certificates from CA
	certpool := x509.NewCertPool()
	ok := certpool.AppendCertsFromPEM(ca)

	if !ok {
		return nil, fmt.Errorf("CA is invalid")
	}

	// Import client certificate/key pair
	_, err := tls.X509KeyPair(certificate, privkey)
	if err != nil {
		return nil, err
	}

	// Create tls.Config with desired tls properties
	return &tls.Config{
		// RootCAs = certs used to verify server cert.
		RootCAs: certpool,
		// ClientAuth = whether to request cert from server.
		// Since the server is set up for SSL, this happens
		// anyways.
		ClientAuth: tls.NoClientCert,
		// ClientCAs = certs used to validate client cert.
		ClientCAs: nil,
		// InsecureSkipVerify = verify that cert contents
		// match server. IP matches what is in cert etc.
		InsecureSkipVerify: false,
		// Certificates = list of certs client sends to server.
		//Certificates: []tls.Certificate{cert},
		Certificates: nil,
	}, nil
}

func (w *Worker) Run(ctx context.Context) {
	verboseLogger.Printf("[%d] initializing\n", w.WorkerId)

	queue := make(chan [2]string)
	cid := w.WorkerId
	_ = randomSource.Int31()

	_, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	topicName := fmt.Sprintf(topicNameTemplate, w.WorkerId)
	subscriberClientId := fmt.Sprintf(subscriberClientIdTemplate, w.WorkerId)
	publisherClientId := fmt.Sprintf(publisherClientIdTemplate, w.WorkerId)
	verboseLogger.Printf("[%d] topic=%s subscriberClientId=%s publisherClientId=%s\n", cid, topicName, subscriberClientId, publisherClientId)
	password := "Replace with token"
	tlsConfig := NewTlsConfig2()
	publisherOptions := mqtt.NewClientOptions().SetClientID(publisherClientId).SetUsername("unused").SetPassword(password).AddBroker(w.BrokerUrl)

	subscriberOptions := mqtt.NewClientOptions().SetClientID(subscriberClientId).SetUsername("unused").SetPassword(password).AddBroker(w.BrokerUrl)
	publisherOptions.SetTLSConfig(tlsConfig)
	subscriberOptions.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		queue <- [2]string{msg.Topic(), string(msg.Payload())}
	})

	if len(w.CA) > 0 || len(w.Key) > 0 {
		tlsConfig, err := NewTLSConfig(w.CA, w.Cert, w.Key)
		if err != nil {
			panic(err)
		}
		publisherOptions.SetTLSConfig(tlsConfig)
	}

	if w.SkipTLSVerification {
		setSkipTLS(publisherOptions)
		setSkipTLS(subscriberOptions)
	}

	
	publisher := mqtt.NewClient(publisherOptions)
	verboseLogger.Printf("[%d] connecting publisher\n", w.WorkerId)
	if token := publisher.Connect(); token.WaitTimeout(w.Timeout) && token.Error() != nil {
		resultChan <- Result{
			WorkerId:     w.WorkerId,
			Event:        ConnectFailedEvent,
			Error:        true,
			ErrorMessage: token.Error(),
		}
		return
	}

	verboseLogger.Printf("[%d] starting control loop %s\n", w.WorkerId, topicName)

	receivedCount := 0
	publishedCount := 0

	t0 := time.Now()
	for i := 0; i < w.NumberOfMessages; i++ {
		text := fmt.Sprintf("{\"id\":%d,\"time\":%d}", messageId, time.Now().UTC().Unix())
		atomic.AddUint64(&messageId, 1)
		token := publisher.Publish(topicName, w.PublisherQoS, w.Retained, text)
		published := token.WaitTimeout(w.Timeout)
		if published {
			publishedCount++
			atomic.AddUint64(&msgCount, 1)
		}
		time.Sleep(w.PauseBetweenMessages)
	}
	publisher.Disconnect(5)

	publishTime := time.Since(t0)
	verboseLogger.Printf("[%d] all messages published\n", w.WorkerId)
	resultChan <- Result{
		WorkerId:          w.WorkerId,
		Event:             CompletedEvent,
		PublishTime:       publishTime,
		ReceiveTime:       time.Since(t0),
		MessagesReceived:  receivedCount,
		MessagesPublished: publishedCount,
	}

	verboseLogger.Printf("[%d] worker finished\n", w.WorkerId)
}
