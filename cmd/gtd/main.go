package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("Hello, World!")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
}
