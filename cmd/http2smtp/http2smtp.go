package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/eexit/http2smtp/internal/env"
	"github.com/eexit/http2smtp/internal/server"
	"github.com/kelseyhightower/envconfig"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	var e env.Bag
	envconfig.MustProcess("", &e)

	server := server.New(e)
	if err := server.Serve(); err != nil {
		panic(err)
	}
	os.Exit(0)
}
