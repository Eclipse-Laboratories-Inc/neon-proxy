package wssubscriber

import (
	"fmt"

	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/config"
	"github.com/neonlabsorg/neon-service-framework/pkg/service"
)

func ServiceHandler(s *service.Service) {
	cfg, err := config.CreateConfigFromEnv()
	if err != nil {
		panic(err)
	}

	subscriber := NewWSSubscriber(
		cfg,
		s.GetContext(),
		s.GetLogger(),
	)

	err = subscriber.Run()
	if err != nil {
		fmt.Println(err)
	}
}
