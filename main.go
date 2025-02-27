package main

import (
	"log"
	"runtime"
	"httpserver/modules"
	"httpserver/modules/server"
)

func main() {

	// setting maximum processers used to prevent any multithreading
	runtime.GOMAXPROCS(1)
	config := modules.LoadConfig("config.json")

	server := server.New(config)
	log.Println(server.Run())
}
