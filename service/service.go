package service

import (
	"context"
	"distributed/registry"
	"fmt"
	"log"
	"net/http"
)

func Start(ctx context.Context, registration registry.Registration, host, port string, registerHandlersFunc func()) (context.Context, error) {
	registerHandlersFunc()
	ctx = startService(ctx, registration.ServiceName, host, port)
	err := registry.DoRegistry(registration)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

func startService(ctx context.Context, serviceName registry.ServiceName, host string, port string) context.Context {
	ctx, cancal := context.WithCancel(ctx)
	var srv http.Server
	srv.Addr = ":" + port

	go func() {
		log.Println(srv.ListenAndServe())
		err := registry.DoShutdown(fmt.Sprintf("http://%s:%s", host, port))
		if err != nil {
			log.Println(err)
		}
		cancal()
	}()

	go func() {
		fmt.Printf("%v started. Press any key to stop \n", serviceName)
		var s string
		fmt.Scanln(&s)
		err := registry.DoShutdown(fmt.Sprintf("http://%s:%s", host, port))
		if err != nil {
			log.Println(err)
		}
		srv.Shutdown(ctx)
		cancal()
	}()
	return ctx
}
