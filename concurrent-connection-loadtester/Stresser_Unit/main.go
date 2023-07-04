package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/ztrue/shutdown"
)

func init() {
	path, err := os.Getwd()
	if err != nil {
		log.Error().Err(err).Msg("")
	}
	log.Info().Err(err).Msg(`path: ` + path)
	viper.SetConfigType(`json`)
	viper.SetConfigName(`config`)
	viper.AddConfigPath(`./`)
	viper.AddConfigPath(`../`)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Error().Err(err).Msg("config file not found")
		}
		panic(err)
	}

	if viper.GetBool(`debug`) {
		log.Info().Msg("Service RUN on DEBUG mode")
	}
}

func NewTlsConfig() *tls.Config {
	certpool := x509.NewCertPool()
	ca := "-----BEGIN CERTIFICATE-----\nMIIDVjCCAj6gAwIBAgITZSbEZ/rEg6U4OYgwH8okbPq1OTANBgkqhkiG9w0BAQsF\nADAyMRUwEwYDVQQKEwxLb3Jld2lyZWxlc3MxGTAXBgNVBAMTEGtvcmV3aXJlbGVz\ncy5jb20wIBcNMjIxMTExMDY0ODA2WhgPMjA1MjExMDMwNjQ4MDVaMDIxFTATBgNV\nBAoTDEtvcmV3aXJlbGVzczEZMBcGA1UEAxMQa29yZXdpcmVsZXNzLmNvbTCCASIw\nDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAIs2ZN6edS0d+UkhQbSJrgC1ZwBi\n+XrumuECOvz9Giwr2nRbsZofjGYAy+PdKfkzlJD6aOWKtX9tx5NWlihbSkGbc8kj\nYQlLwrm/6/gZWHjZeMY8rj18+ieIzFB/Y3sK3LmkrGnmpw1FQJtTDoOf0S6YfWIX\ngRv1qJrZevW7nrzlzzex3dmJroj9jAcyxgj+VK7IrBfUJTq4vQ4w6ltPHKh9ZxNL\nyIaHb94BUpPugXecwAyZuKjEFPH8z62bwo9uSnJwogshFueIAF9Nw57J+UgsCFrY\nqJHLz4DSYlhZhAaqoYDCDUIaz5xQkW0ggywqOYrkB7RK+r1W6c8MIgF4JI8CAwEA\nAaNjMGEwDgYDVR0PAQH/BAQDAgEGMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYE\nFBm2Uovs4uqrVGW7OkUmnPXv3BBFMB8GA1UdIwQYMBaAFBm2Uovs4uqrVGW7OkUm\nnPXv3BBFMA0GCSqGSIb3DQEBCwUAA4IBAQBFBeKwG6l528TjrGjNlGumkBjFaPK0\n5GS5/MWBEqOzXRzA1zs2XpEl4/HH7LdOtJvI2S+oSkbOYOVF3hJoh7H/Z/jOsbDk\ns/OXu+BDPnJN6RgmtcqDAZa/KtKLhcIGCwK5Sl/C+Vx6bUogXTCai8RwnGB9XwF1\n6EUZcyaaZpcJg0wEjYUm/tyvehnpR8Usl9aDUdpKM0NgTgCdWaAQCNT4HpTvHe1v\n8R97c9OpcTBc/Cs2XadBtOY2lwDHskMij50+n+xzas+jBPbrnTbYIjcixSEiSsxm\n5vVHrDWO9naPgbO7sIuExvhJSBMeBeQBQ4wUFma5Endcg4ZE8CCqGlsn\n-----END CERTIFICATE-----\n"
	bytesCa := []byte(ca)
	certpool.AppendCertsFromPEM(bytesCa)
	return &tls.Config{
		RootCAs: certpool,
	}
}

type responseHttp struct {
	ClientStart uint64 `json:"clientStart" validate:""`
	TimeStart   uint64 `json:"timeStart" validate:"required"`
}

func getTokenValue(url string) (uint64, uint64) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	var marshaledBody responseHttp
	json.Unmarshal(body, &marshaledBody)
	return marshaledBody.ClientStart, marshaledBody.TimeStart

}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func connectMqttClient(url string, id uint64, channel chan struct{}) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(url)
	opts.SetAutoReconnect(true)
	clientId := fmt.Sprintf("subscriptions/REPLACE WITH SUBSCRIPTIONID/registries/Replace With Registry ID/devices/Stresser%d", id)
	opts.SetClientID(clientId)
	opts.SetPingTimeout(10 * time.Minute)
	opts.SetProtocolVersion(4)
	opts.SetUsername("unused")
	opts.SetPassword("REPLACE WITH TOKEN")
	opts.SetTLSConfig(NewTlsConfig())
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.SetConnectRetryInterval(30 * time.Second)
	opts.SetKeepAlive(300)
	opts.SetOnConnectHandler(connectHandler)
	opts.SetConnectionLostHandler(connectLostHandler)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Error().Err(token.Error()).Msg("")
	}
	<-channel
	client.Disconnect(100)
	log.Print(opts.ClientID + " disconnected")

}

func main() {
	log.Info().Msg("Go Time")
	flag.Parse()

	viper.AutomaticEnv()
	viper.SetEnvPrefix(viper.GetString("ENV"))
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	mqttUrl := viper.GetString("ENV_MQTT_URL")
	if mqttUrl == "" {
		log.Fatal().Msg("Configuration Error: MQTT URL  not available")
	}
	tokenUrl := viper.GetString("ENV_TOKEN_URL")
	if tokenUrl == "" {
		log.Fatal().Msg("Configuration Error: Token URL  not available")
	}
	maxClients := viper.GetUint64("ENV_MAX_CLIENTS")
	if maxClients == 0 {
		log.Fatal().Msg("Configuration Error: Max Clients  not available")
	}
	token, timeStart := getTokenValue(tokenUrl)
	log.Print(token)
	log.Print(timeStart)
	token = token * maxClients
	time.Sleep(time.Duration(timeStart) * time.Second)
	log.Print("starting stress")
	loops := token + maxClients
	log.Print(loops)
	disconnectChannel := make(chan struct{}, maxClients)

	for i := token; i < loops; i++ {
		abc := i
		time.Sleep(50 * time.Microsecond)
		go func() {
			connectMqttClient(mqttUrl, abc, disconnectChannel)
		}()
	}

	log.Print("clients connected")
	shutdown.Add(func() {
		var i uint64
		for i = 0; i < maxClients; i++ {
			disconnectChannel <- struct{}{}
		}
		log.Info().Msg("Stopping...")
		time.Sleep(5 * time.Second)
		log.Info().Msg("Stopped")
	})
	shutdown.Listen(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
}
