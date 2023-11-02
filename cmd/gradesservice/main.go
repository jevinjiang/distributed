package main

import (
	"context"
	"distributed/grades"
	"distributed/log"
	"distributed/registry"
	"distributed/service"
	"fmt"
	stlog "log"
)

func main() {
	host, port := "localhost", "6000"
	serviceAddress := fmt.Sprintf("http://%s:%s", host, port)
	r := registry.Registration{
		ServiceName:      registry.GradingService,
		ServiceURL:       serviceAddress,
		RequiredServices: []registry.ServiceName{registry.LogService},
		ServiceUpdateURL: serviceAddress + "/services",
		HeartbeatUrl:     serviceAddress + "/heartbeat",
	}
	ctx, err := service.Start(
		context.Background(),
		r,
		host,
		port,
		grades.RegisterHandlers,
	)
	if err != nil {
		stlog.Fatal(err)
	}
	if logProvider, err := registry.GetProvider(registry.LogService); err == nil {
		fmt.Printf("logging servide found at %s\n", logProvider)
		log.SetClientLogger(logProvider, r.ServiceName)
	}
	<-ctx.Done()
	fmt.Printf("Shutting down grade service.")
}
