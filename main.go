package main

import (
	"fmt"
	"os"

	"MineMock/internal/config"
	"MineMock/internal/server"
)

func main() {
	cfg := config.FromEnv()
	addr := cfg.Address()

	if err := server.Run(addr, cfg.ErrorMessage); err != nil {
		fmt.Println("Server error:", err)
		os.Exit(1)
	}
}
