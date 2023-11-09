package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

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

	client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("* [%s] %s\n", msg.Topic(), string(msg.Payload()))
	})
}

func main() {

	topic := "foo"

	go mqttListen(topic)

	client := mqttConnect("go-server")

	http.HandleFunc("/send-message", func(w http.ResponseWriter, r *http.Request) {
		sendMessage(w, r, client)
	})

	http.Handle("/", http.FileServer(http.Dir("./static")))

	log.Println("Listening on port 8000...")

	log.Fatal(http.ListenAndServe(":8000", nil))

}

func sendMessage(w http.ResponseWriter, r *http.Request, client mqtt.Client) {

	// defer r.Body.Close()

	// message, err := io.ReadAll(r.Body)

	// if err != nil {
	// 	log.Fatalln(err)
	// }

	r.ParseForm()

	message := r.Form.Get("message")

	topic := "foo"

	client.Publish(topic, 0, false, message)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Message Sent!"))
}
