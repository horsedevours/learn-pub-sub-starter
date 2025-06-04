package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	gamelogic.PrintServerHelp()

	const connStr = "amqp://guest:guest@localhost:5672/"

	fmt.Println("Starting Peril server...")

	conn, err := amqp.Dial(connStr)
	if err != nil {
		log.Fatalf("error connecting to rabbit: %s", err)
	}
	defer conn.Close()

	fmt.Println("Connection established...")

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("error creating channel: %s", err)
	}
	defer channel.Close()

	err = pubsub.SubscribeGob(conn, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+".*", pubsub.Durable, handlerGameLogs())
	if err != nil {
		log.Fatalf("error subscribing to game logs queue: %s", err)
	}

	for {
		input := gamelogic.GetInput()
		if len(input) < 1 {
			continue
		}
		if input[0] == "pause" {
			fmt.Println("Pausing game...")
			pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			})
			continue
		}
		if input[0] == "resume" {
			fmt.Println("Resuming game...")
			pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			})
			continue
		}
		if input[0] == "quit" {
			fmt.Println("Exiting game...")
		}
	}
}
