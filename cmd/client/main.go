package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

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

	channel, _, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, fmt.Sprintf("%s.%s", routing.PauseKey, user), routing.PauseKey, pubsub.Transient)
	if err != nil {
		log.Fatalf("error binding queue: %s", err)
	}

	gameState := gamelogic.NewGameState(user)

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, fmt.Sprintf("%s.%s", routing.PauseKey, user), routing.PauseKey, pubsub.Transient, handlerPause(gameState))
	if err != nil {
		log.Fatalln("error subscribing to pause handler: ", err)
	}

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, user), fmt.Sprintf("%s.*", routing.ArmyMovesPrefix),
		pubsub.Transient, handlerArmyMove(channel, gameState))
	if err != nil {
		log.Fatalln("error subscribing to army moves handler: ", err)
	}

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix, fmt.Sprintf("%s.*", routing.WarRecognitionsPrefix),
		pubsub.Durable, handlerWarMoves(channel, gameState))
	if err != nil {
		log.Fatalln("error subscribing to army moves handler: ", err)
	}

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
			mv, err := gameState.CommandMove(words)
			if err != nil {
				log.Printf("error moving units: %s\n", err)
			}
			err = pubsub.PublishJSON(channel, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, user), mv)
			if err != nil {
				log.Println("error publishing move: ", err)
			}
			log.Printf("move published successfully")
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
			if len(words) < 2 {
				log.Println("You forgot to tell me how much spam!")
				continue
			}
			count, err := strconv.Atoi(words[1])
			if err != nil {
				log.Println("error converting argument: ", err)
				continue
			}
			for i := range count {
				err = pubsub.PublishGob(channel, routing.ExchangePerilTopic, routing.GameLogSlug+"."+user, routing.GameLog{
					CurrentTime: time.Now(),
					Message:     gamelogic.GetMaliciousLog(),
					Username:    user,
				})
				if err != nil {
					log.Println("error publishing spam #", i)
				}
			}
			continue
		}
		if words[0] == "quit" {
			gamelogic.PrintQuit()
			break
		}
	}

}
