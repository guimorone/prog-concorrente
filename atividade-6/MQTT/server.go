package main

import (
	"fmt"
	"strconv"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const broker = "tcp://localhost:1883"
const clientID = "server"

func main() {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	defer client.Disconnect(250)

	client.Subscribe("tirequality", 0, func(client MQTT.Client, msg MQTT.Message) {
		newTireQualityAux, err := strconv.ParseFloat(string(msg.Payload()), 32)
		if err != nil {
			fmt.Println("Error to convert value:", err)
			return
		}
		
		newTireQuality := fmt.Sprintf("%.2f", (newTireQualityAux*0.97))
		fmt.Printf("%s\n", newTireQuality)
		client.Publish("newtirequality", 0, false, newTireQuality)
	})

	select {}
}
