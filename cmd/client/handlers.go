package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/rabbitmq"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) rabbitmq.AckType {
	return func(ps routing.PlayingState) rabbitmq.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)

		return rabbitmq.Ack
	}
}

func handlerMove(c *rabbitmq.Client, gs *gamelogic.GameState) func(gamelogic.ArmyMove) rabbitmq.AckType {
	return func(move gamelogic.ArmyMove) rabbitmq.AckType {
		defer fmt.Print("> ")

		switch gs.HandleMove(move) {
		case gamelogic.MoveOutComeSafe:
			return rabbitmq.Ack

		case gamelogic.MoveOutcomeMakeWar:
			if err := c.Publish(routing.ExchangePerilTopic, routing.WarRecognitionsPrefix+"."+gs.GetUsername(), gamelogic.RecognitionOfWar{Attacker: move.Player, Defender: gs.GetPlayerSnap()}); err != nil {
				return rabbitmq.NackRequeue
			}
			return rabbitmq.Ack

		case gamelogic.MoveOutcomeSamePlayer:
			fallthrough
		default:
			return rabbitmq.NackDiscard
		}
	}
}

func handlerWar(gs *gamelogic.GameState) func(rw gamelogic.RecognitionOfWar) rabbitmq.AckType {
	return func(rw gamelogic.RecognitionOfWar) rabbitmq.AckType {
		defer fmt.Print("> ")
		warOutcome, _, _ := gs.HandleWar(rw)

		switch warOutcome {
		case gamelogic.WarOutcomeNotInvolved:
			return rabbitmq.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return rabbitmq.NackDiscard
		case gamelogic.WarOutcomeYouWon:
			fallthrough
		case gamelogic.WarOutcomeDraw:
			fallthrough
		case gamelogic.WarOutcomeOpponentWon:
			return rabbitmq.Ack
		default:
			log.Println("Unknown war outcome")
			return rabbitmq.NackDiscard
		}
	}
}
