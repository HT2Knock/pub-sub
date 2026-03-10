package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/rabbitmq"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

type config struct {
	addr string
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Args, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	log.Println("shutdown complete")
	os.Exit(0)
}

func run(ctx context.Context, args []string, stdout io.Writer) error {
	var cfg config

	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(stdout)

	flags.StringVar(&cfg.addr, "addr", "amqp://guest:guest@localhost:5672", "addr of RabbitMQ amqp")

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	client, err := rabbitmq.NewClient(cfg.addr)
	if err != nil {
		return err
	}
	defer client.Close()

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		return err
	}

	gs := gamelogic.NewGameState(username)

	if err = rabbitmq.Subscribe(
		*client,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		rabbitmq.Transient,
		HandlePause(gs)); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		inputs := gamelogic.GetInput()
		if len(inputs) < 1 {
			continue
		}

		switch inputs[0] {
		case "spawn":
			if err := gs.CommandSpawn(inputs); err != nil {
				log.Printf("Spawn failed: %v\n", err)
			}
		case "move":
			mv, err := gs.CommandMove(inputs)
			if err != nil {
				log.Printf("Move failed: %v\n", err)
			}

			fmt.Printf("%v move a unit\n", mv.Player.Username)

		case "status":
			gs.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			fmt.Println("Spamming is not allowed yet!")

		case "quit":
			log.Println("quitting good bye...")
			return nil

		default:
			fmt.Println("unknown command!")
		}
	}
}

func HandlePause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	}
}
