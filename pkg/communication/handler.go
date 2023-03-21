package communication

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/neonlabsorg/neon-proxy/internal/server"
	"github.com/neonlabsorg/neon-proxy/pkg/gRPC"
	"github.com/neonlabsorg/neon-proxy/pkg/service/configuration"
)

type handler struct {
	client  gRPC.EventClient
	context context.Context
}

func newHandler(c gRPC.EventClient, ctx context.Context) *handler {
	return &handler{
		client:  c,
		context: ctx,
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// implement GET method
	if r.Method == "GET" {
		var roleName string
		if names, ok := r.URL.Query()["name"]; ok {
			roleName = names[0]
		}

		instances, err := h.client.GetInstances(h.context, &gRPC.RoleData{
			Role: int32(configuration.FromString(roleName)),
		})
		if err != nil {
			log.Println("could not get Instances", err)
		}

		items := make([]server.Instance, 0, len(instances.Items))
		for _, inst := range instances.Items {
			items = append(items, server.Instance{
				Role:      inst.GetRole(),
				Ip:        inst.GetIp(),
				ID:        inst.GetId(),
				CreatedAt: inst.GetCreatedAt().AsTime(),
			})
		}

		r := Instances{Items: items}
		json.NewEncoder(w).Encode(r)
	}
}

type Instances struct {
	Items []server.Instance `json:"instances,omitempty"`
}
