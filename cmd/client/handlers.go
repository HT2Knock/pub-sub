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
			if err := c.PublishJSON(routing.ExchangePerilTopic, routing.WarRecognitionsPrefix+"."+gs.GetUsername(), gamelogic.RecognitionOfWar{Attacker: move.Player, Defender: gs.GetPlayerSnap()}); err != nil {
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

func handlerWar(client *rabbitmq.Client, gs *gamelogic.GameState) func(rw gamelogic.RecognitionOfWar) rabbitmq.AckType {
	return func(rw gamelogic.RecognitionOfWar) rabbitmq.AckType {
		defer fmt.Print("> ")
		warOutcome, w, l := gs.HandleWar(rw)

		logs := func(msg string) rabbitmq.AckType {
			if err := publishGameLog(client, gs.Player.Username, msg); err != nil {
				return rabbitmq.NackRequeue
			}

			return rabbitmq.Ack
		}

		switch warOutcome {
		case gamelogic.WarOutcomeNotInvolved:
			return rabbitmq.NackRequeue

		case gamelogic.WarOutcomeNoUnits:
			return rabbitmq.NackDiscard

		case gamelogic.WarOutcomeDraw:
			return logs(fmt.Sprintf("A war between %s and %s resulted in a draw", w, l))

		case gamelogic.WarOutcomeYouWon:
			fallthrough

		case gamelogic.WarOutcomeOpponentWon:
			return logs(w + " won a war against " + l)

		default:
			log.Println("Unknown war outcome")
		}

		return rabbitmq.Ack
	}
}
