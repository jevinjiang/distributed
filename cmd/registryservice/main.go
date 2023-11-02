package main

import (
	"context"
	"distributed/registry"
	"fmt"
	"log"
	"net/http"
)

func main() {
	registry.SetupRegistryService()
	http.Handle("/services", &registry.RegistryService{})
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	var srv http.Server
	srv.Addr = registry.ServerPort

	go func() {
		log.Println(srv.ListenAndServe())
		cancelFunc()
	}()

	go func() {
		fmt.Println("Register service started.Press any key to stop.")
		var s string
		fmt.Scanln(&s)
		srv.Shutdown(ctx)
		cancelFunc()
	}()

	<-ctx.Done()
	fmt.Println("Shutting down register service")
}
