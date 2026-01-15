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

	"github.com/rabbitmq/amqp091-go"
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

	conn, err := amqp091.Dial(cfg.addr)
	if err != nil {
		return fmt.Errorf("dialing rabbitmq: %w", err)
	}
	defer func() {
		log.Println("closing rabbitmq connection...")
		conn.Close()
	}()

	log.Println("RabbitMQ connnect successfully")

	<-ctx.Done()

	return nil
}
