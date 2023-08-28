package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const broker = "tcp://localhost:1883"

var clientID = strconv.FormatInt(time.Now().UnixNano(), 10)

func main() {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	rttTimes := make([]time.Duration, 10001)
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	defer client.Disconnect(250)
	var newTireQuality string
	tireQuality:= 40+ rand.Float64()*(60)
	for i := 0; i < 10000; i++ {
		start := time.Now()
		payload := fmt.Sprintf("%f", tireQuality)
		token := client.Publish("tirequality", 0, false, payload)
		token.Wait()
		client.Subscribe("newtirequality", 0, func(client MQTT.Client, msg MQTT.Message) {
			newTireQuality = string(msg.Payload())
			//fmt.Printf("opa %s",newTireQuality)
			newTireQualityAux, err := strconv.ParseFloat(newTireQuality, 32)
			//fmt.Printf("resposta %s\n", newTireQualityAux)
			if err != nil {
				fmt.Println(err)
			} else {
				if (newTireQualityAux < 20) {
					tireQuality = 100.0
				} else {
					tireQuality = newTireQualityAux	
				}
				timeElapsed := time.Since(start)
				rttTimes[i] = timeElapsed
			}
		})
		
	}
	var totalRTT time.Duration
	for _, rtt := range rttTimes {
		totalRTT += rtt
	}
	averageRTT := totalRTT / time.Duration(10001)
	fmt.Println("Tempo médio de RTT:", averageRTT)
}

