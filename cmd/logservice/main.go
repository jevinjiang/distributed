package main

import (
	"context"
	"distributed/log"
	"distributed/registry"
	"distributed/service"
	"fmt"
	stlog "log"
)

func main() {
	log.Run("./distributed.log")
	host, port := "localhost", "4000"
	serviceAddress := fmt.Sprintf("http://%s:%s", host, port)
	r := registry.Registration{
		ServiceName:      registry.LogService,
		ServiceURL:       serviceAddress,
		RequiredServices: make([]registry.ServiceName, 0),
		ServiceUpdateURL: serviceAddress + "/services",
		HeartbeatUrl:     serviceAddress + "/heartbeat",
	}
	ctx, err := service.Start(
		context.Background(),
		r,
		host,
		port,
		log.RegisterHandler,
	)
	if err != nil {
		stlog.Fatal(err)
	}
	<-ctx.Done()
	fmt.Printf("Shutting down log service.")
}
