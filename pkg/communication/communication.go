package communication

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/neonlabsorg/neon-proxy/internal/server"
	"github.com/neonlabsorg/neon-proxy/pkg/gRPC"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	"github.com/neonlabsorg/neon-proxy/pkg/service/configuration"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CommunicationServer struct {
	ctx                             context.Context
	role                            configuration.Role
	instanceID                      string
	ip                              string
	cluster                         string
	services                        map[configuration.Role][]configuration.ProtocolConfiguration
	communicationServerPort         string
	communicationEndpointServerPort string
	clientConnection                *grpc.ClientConn
	eventClients                    []gRPC.EventClient
}

func NewCommunicationServer(
	ctx context.Context,
	role configuration.Role,
	instanceID string,
	ip string,
	cluster string,
	ss map[configuration.Role][]configuration.ProtocolConfiguration,
	communicationServerPort string,
	communicationEndpointServerPort string,
) *CommunicationServer {
	return &CommunicationServer{
		ctx:                             ctx,
		role:                            role,
		instanceID:                      instanceID,
		ip:                              ip,
		cluster:                         cluster,
		services:                        ss,
		communicationServerPort:         communicationServerPort,
		communicationEndpointServerPort: communicationEndpointServerPort,
	}
}

func (s *CommunicationServer) Shutdown() {
	s.clientConnection.Close()
}

func (s *CommunicationServer) Init() error {
	var err error
	s.clientConnection, err = grpc.Dial(s.communicationServerPort, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(3*time.Second))
	if err != nil {
		return err
	}

	for _, configs := range s.services {
		for i := 0; i < len(configs); i++ {
			eventClient := gRPC.NewEventClient(s.clientConnection)
			s.eventClients = append(s.eventClients, eventClient)
		}
	}

	for _, cli := range s.eventClients {
		res, err := cli.AfterCreate(s.ctx, &gRPC.OnCreate{Instance: &gRPC.Instance{
			Role:      int32(s.role),
			Id:        s.instanceID,
			Ip:        s.ip,
			Cluster:   s.cluster,
			CreatedAt: timestamppb.Now(),
		}})
		if err != nil {
			s.Shutdown()
			return err
		}
		if !res.Success {
			logger.Debug().Msgf("could not add created instance into collector %s, %s, %s", s.cluster, s.instanceID, s.ip)
		}
	}

	go func() {
		tick := time.NewTicker(time.Second * time.Duration(3*time.Second))
		for {
			select {
			case <-s.ctx.Done():
				tick.Stop()
				s.Shutdown()
				return
			case <-tick.C:
				s.processHelthChecker()
			}
		}
	}()

	return nil
}

func (s *CommunicationServer) processHelthChecker() {
	for _, cli := range s.eventClients {
		res, err := cli.HealthCheck(s.ctx, &gRPC.OnHealthCheck{Instance: &gRPC.Instance{
			Role:      int32(s.role),
			Id:        s.instanceID,
			Ip:        s.ip,
			Cluster:   s.cluster,
			CreatedAt: timestamppb.Now(),
		}})

		if err != nil || !res.Success {
			// todo(argishti) should we remove client connection?
			logger.Debug().Msgf("could not update instance into collector %s %s, %s", s.cluster, s.instanceID, s.ip)
		}
	}
}

func (s *CommunicationServer) RegisterServer() error {
	// tcp listen to the port
	lis, err := net.Listen("tcp", s.communicationServerPort)
	// check for errors
	if err != nil {
		return fmt.Errorf("failed to listen port: %w", err)
	}
	// instantiate the server
	srv := grpc.NewServer()

	// register server method (actions the server will do)
	gRPC.RegisterEventServer(srv, &server.EventServer{})

	if err := srv.Serve(lis); err != nil {
		return err
	}
	return nil
}

func (s *CommunicationServer) RunServer() error {
	cli, err := grpc.Dial(s.communicationServerPort, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(3*time.Second))
	if err != nil {
		s.Shutdown()
		return err
	}

	client := gRPC.NewEventClient(cli)
	servicesHandler := newHandler(client, s.ctx)
	http.Handle("/services", servicesHandler)

	srv := &http.Server{Addr: s.communicationEndpointServerPort}

	go func() {
		defer cli.Close()
		<-s.ctx.Done()

		if err := srv.Shutdown(context.Background()); err != nil {
			logger.Error().Err(err).Msg("could not shutdown communication server")
		}
	}()
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}
