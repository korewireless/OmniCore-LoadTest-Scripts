# MQTT Stresser

Load testing tool to stress MQTT message broker

Based ON https://github.com/inovex/mqtt-stresser


# Steps

1. Create many devices(For example 300) with any registry and Suffix of device name as Stresser,ex:Stresser0,Stresser1 etc.Use a single key certificate file so its easier to manage.Generate a token for 10 hours and replace it in worker.go line 136.Also replace the registry id in main.go line 145 146.

2. In Root Folder of throughpout-stresser,Run
 `go run . -broker ssl://replacewithsubUniquestring.mqtt.korewireless.com:8883 -num-clients 300 -num-messages 1000 -rampup-size 10000 -publisher-qos 1 -subid ReplacewithsubscriptionId -pause-between-messages 0.25s`

EXAMPLE:
 `go run . -broker ssl://k7x0roxvg5.mqtt.korewireless.com:8883 -num-clients 300 -num-messages 1000 -rampup-size 10000 -publisher-qos 1 -subid korewireless-development -pause-between-messages 0.25s`

## Args
   - subid: "Subscription name of kore client")
   - num-clients: "Number of concurrent clients")
   - num-messages: "Number of messages shipped by client")
   - rampup-size: "Size of rampup batch. Default rampup batch size is 100.")
   - broker: "Broker URL"        
   - publisher-qos" "QoS level of published messages")
   - pause-between-messages" "Adds a pause between sending messages to simulate sensors sending messages infrequently")
