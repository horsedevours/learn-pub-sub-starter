package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	winnerLoser = "%s won a war against %s"
	draw        = "A war between %s and %s resulted in a draw"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerArmyMove(channel *amqp.Channel, gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(mv gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(mv)
		switch outcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(channel, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.WarRecognitionsPrefix, gs.GetUsername()), gamelogic.RecognitionOfWar{
				Attacker: mv.Player,
				Defender: gs.GetPlayerSnap(),
			})
			if err != nil {
				log.Println("error publishing war move: ", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			return pubsub.NackDiscard
		}
	}
}

func handlerWarMoves(channel *amqp.Channel, gs *gamelogic.GameState) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(rw)
		switch outcome {
		case gamelogic.WarOutcomeOpponentWon:
			err := publishWarLog(channel, gs.GetUsername(), fmt.Sprintf(winnerLoser, winner, loser))
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			err := publishWarLog(channel, gs.GetUsername(), fmt.Sprintf(winnerLoser, winner, loser))
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			err := publishWarLog(channel, gs.GetUsername(), fmt.Sprintf(draw, winner, loser))
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		default:
			log.Println("Unknown outcome value: ", outcome)
			return pubsub.NackDiscard
		}
	}
}

func publishWarLog(channel *amqp.Channel, username, message string) error {
	err := pubsub.PublishGob(channel, routing.ExchangePerilTopic, routing.GameLogSlug+"."+username, routing.GameLog{
		CurrentTime: time.Now(),
		Message:     message,
		Username:    username,
	})
	if err != nil {
		return err
	}
	return nil
}
