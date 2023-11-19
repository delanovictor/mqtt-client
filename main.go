package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type State struct {
	temperature string
	led1        bool
	led2        bool
}

var state State

func mqttConnect(clientId string) mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", "localhost:1883"))
	opts.SetClientID(clientId)

	client := mqtt.NewClient(opts)

	token := client.Connect()

	for !token.WaitTimeout(3 * time.Second) {
	}

	if err := token.Error(); err != nil {
		log.Fatal(err)
	}

	return client
}

func mqttListen(topic string) {
	client := mqttConnect("sub")

	client.Subscribe(topic, 0, func(_ mqtt.Client, msg mqtt.Message) {
		fmt.Printf("* [%s] %s\n", msg.Topic(), string(msg.Payload()))
		state.temperature = string(msg.Payload())
	})
}

func main() {
	go mqttListen("delano_e_mauro")

	client := mqttConnect("go-server")

	http.HandleFunc("/send-message", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		message := r.FormValue("message")

		fmt.Printf("Message: %v\n", message)

		sendMessage(client, message)

		w.Write([]byte("Message Received!"))
	})

	http.HandleFunc("/led1", func(w http.ResponseWriter, _ *http.Request) {
		message := ""
		if state.led1 {
			message = "0"
		} else {
			message = "1"
		}
		state.led1 = !state.led1

		fmt.Printf("LED 1 - New State: %t\n", state.led1)

		sendMessage(client, message)

		fmt.Fprintf(w, `<div id="led1" class="%t"> %t </div>`, state.led1, state.led1)
	})

	http.HandleFunc("/led2", func(w http.ResponseWriter, _ *http.Request) {
		message := ""
		if state.led2 {
			message = "2"
		} else {
			message = "3"
		}

		state.led2 = !state.led2

		fmt.Printf("LED 2 - New State: %t\n", state.led2)

		sendMessage(client, message)

		fmt.Fprintf(w, `<div id="led2" class="%t"> %t </div>`, state.led2, state.led2)
	})

	http.HandleFunc("/temperature", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Println("Temperature: ", state.temperature)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(state.temperature))
	})

	http.Handle("/", http.FileServer(http.Dir("./static")))

	log.Println("Listening on port 8080...")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func sendMessage(client mqtt.Client, message string) {
	topic := "delano_e_mauro2"

	r := client.Publish(topic, 0, false, message)

	fmt.Println(r)
}
