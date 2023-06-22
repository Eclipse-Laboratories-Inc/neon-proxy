package main

import (
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber"
	"github.com/neonlabsorg/neon-service-framework/pkg/service"
	"github.com/neonlabsorg/neon-service-framework/pkg/service/configuration"
)

func main() {
	s := service.CreateService(&configuration.Config{
		Name: "wssubscriber",
	})

	s.AddHandler(wssubscriber.ServiceHandler)

	s.Run()
}
