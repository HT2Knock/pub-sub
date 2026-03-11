package main

import (
	"fmt"

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

func handlerMove(gs *gamelogic.GameState) func(move gamelogic.ArmyMove) rabbitmq.AckType {
	return func(move gamelogic.ArmyMove) rabbitmq.AckType {
		defer fmt.Print("> ")

		switch gs.HandleMove(move) {
		case gamelogic.MoveOutComeSafe:
		case gamelogic.MoveOutcomeMakeWar:
			return rabbitmq.Ack

		case gamelogic.MoveOutcomeSamePlayer:
		default:
			return rabbitmq.NackDiscard
		}

		return rabbitmq.NackDiscard
	}
}
