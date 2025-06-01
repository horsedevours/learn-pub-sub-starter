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
	const connStr = "amqp://guest:guest@localhost:5672/"
	fmt.Println("Starting Peril client...")

	conn, err := amqp.Dial(connStr)
	if err != nil {
		log.Fatalf("error connecting to rabbit: %s", err)
	}
	defer conn.Close()

	fmt.Println("Connection established...")

	user, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("error with user login: %s", err)
	}

	_, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, fmt.Sprintf("%s.%s", routing.PauseKey, user), routing.PauseKey, int(pubsub.Transient))
	if err != nil {
		log.Fatalf("error binding queue: %s", err)
	}

	gameState := gamelogic.NewGameState(user)

	for !conn.IsClosed() {
		words := gamelogic.GetInput()

		if words[0] == "spawn" {
			err = gameState.CommandSpawn(words)
			if err != nil {
				log.Printf("error spawning units: %s\n", err)
			}
			continue
		}
		if words[0] == "move" {
			_, err := gameState.CommandMove(words)
			if err != nil {
				log.Printf("error moving units: %s\n", err)
			}
			fmt.Println("Move successful!")
			continue
		}
		if words[0] == "status" {
			gameState.CommandStatus()
			continue
		}
		if words[0] == "help" {
			gamelogic.PrintClientHelp()
			continue
		}
		if words[0] == "spam" {
			fmt.Println("Spamming not allowed yet!")
			continue
		}
		if words[0] == "quit" {
			gamelogic.PrintQuit()
			break
		}
	}

}
