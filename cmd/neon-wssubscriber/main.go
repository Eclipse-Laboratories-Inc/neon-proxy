package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/neonlabsorg/neon-proxy/pkg/service"
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber"
	"github.com/neonlabsorg/neon-proxy/pkg/service/configuration"
)

func main() {
	fmt.Println(runtime.NumGoroutine())
	time.Sleep(time.Second * 3)
	fmt.Println(runtime.NumGoroutine())
	s := service.CreateService(&configuration.Config{
		Name: "wssubscriber",
		// Storage: &configuration.ConfigStorageList{},
	})

	s.AddHandler(runWSSubscriberProxy)

	s.Run()
}

func runWSSubscriberProxy(s *service.Service) {
	cfg, err := wssubscriber.CreateConfigFromEnv()
	if err != nil {
		panic(err)
	}

	subscriber := wssubscriber.NewWSSubscriber(
		cfg,
		s.GetContext(),
		s.GetLogger(),
	)

	err = subscriber.Run()
	if err != nil {
		fmt.Println(err)
	}
}
