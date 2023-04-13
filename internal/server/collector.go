package server

import (
	"context"
	"sync"
	"time"

	"github.com/neonlabsorg/neon-proxy/pkg/gRPC"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	once               sync.Once
	collectoreInstance *collector
)

const HealthCheckDelta = 3 * time.Second

// Instance type
type Instance struct {
	Role      int32     `json:"role,omitempty"`
	Ip        string    `json:"ip,omitempty"`
	ID        string    `json:"id,omitempty"`
	Cluster   string    `json:"clusterName,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

// Instance Key attributes
type InstanceKey struct {
	Role    int32
	ip      string
	id      string
	cluster string
}

// Uptodate Instance Value type
type InstanceValue struct {
	createdAt time.Time
	updatedAt time.Time
}

// convert gRPC Instance to InstanceKey
func toInstanceKey(i *gRPC.Instance) (InstanceKey, time.Time) {
	return InstanceKey{
		Role:    i.GetRole(),
		ip:      i.GetIp(),
		id:      i.GetId(),
		cluster: i.GetCluster(),
	}, i.CreatedAt.AsTime()
}

// converts InstanceKey to Instance
func toInstance(rk InstanceKey, t time.Time) Instance {
	return Instance{
		Role:      rk.Role,
		Ip:        rk.ip,
		ID:        rk.id,
		Cluster:   rk.cluster,
		CreatedAt: t,
	}
}

// defince up to date service items
type upToDateServices struct {
	items map[InstanceKey]InstanceValue
}

// returns up to date services
func (up *upToDateServices) Items() map[InstanceKey]InstanceValue {
	now := time.Now()
	for instance, ts := range up.items {
		updatedAt := ts.updatedAt
		if now.Sub(updatedAt) >= HealthCheckDelta {
			delete(up.items, instance)
		}
	}
	return up.items
}

// type for collecting Instances
type collector struct {
	mx       sync.Mutex
	services upToDateServices
}

// adds Instance
func (c *collector) AddInstance(r *gRPC.Instance) bool {
	c.mx.Lock()
	defer c.mx.Unlock()

	instanceKey, t := toInstanceKey(r)
	services := c.services.Items()
	_, exist := services[instanceKey]
	services[instanceKey] = InstanceValue{createdAt: t, updatedAt: time.Now()}
	return !exist
}

// removes Instance
func (c *collector) RemoveInstance(r *gRPC.Instance) bool {
	c.mx.Lock()
	defer c.mx.Unlock()

	instance, _ := toInstanceKey(r)
	services := c.services.Items()
	_, exist := services[instance]
	delete(services, instance)
	return exist
}

// updates Instance
func (c *collector) UpdateInstanceTimestamp(r *gRPC.Instance) bool {
	c.mx.Lock()
	defer c.mx.Unlock()

	instance, t := toInstanceKey(r)
	services := c.services.Items()
	_, exist := services[instance]
	if exist {
		services[instance] = InstanceValue{createdAt: t, updatedAt: time.Now()}
	}
	return exist
}

// returns instancess by name
func (c *collector) GetInstances(role int32) []Instance {
	c.mx.Lock()
	defer c.mx.Unlock()

	var instances []Instance
	for instanceKey, t := range c.services.Items() {
		if role == instanceKey.Role {
			instances = append(instances, toInstance(instanceKey, t.createdAt))
		}
	}
	return instances
}

// returns all instancess from collector
func (c *collector) GetAllInstances() []Instance {
	c.mx.Lock()
	defer c.mx.Unlock()

	services := c.services.Items()
	insts := make([]Instance, 0, len(services))
	for instanceKey, t := range services {
		insts = append(insts, toInstance(instanceKey, t.createdAt))
	}
	return insts
}

// returns singletone instance
func getCollector() *collector {
	once.Do(func() {
		collectoreInstance = &collector{
			services: upToDateServices{items: make(map[InstanceKey]InstanceValue)},
		}
	})
	return collectoreInstance
}

// implement gRPC.EventClient interface
type EventServer struct {
	gRPC.UnimplementedEventServer
}

// implement the gRPC.EventClient interface
func (ev *EventServer) AfterCreate(_ context.Context, oc *gRPC.OnCreate) (*gRPC.Response, error) {
	success := getCollector().AddInstance(oc.Instance)
	return &gRPC.Response{Success: success}, nil
}

// implement the gRPC.EventClient interface
func (ev *EventServer) BeforeShutDown(_ context.Context, sd *gRPC.OnShutDown) (*gRPC.Response, error) {
	success := getCollector().RemoveInstance(sd.Instance)
	return &gRPC.Response{Success: success}, nil
}

// implement the gRPC.EventClient interface
func (ev *EventServer) HealthCheck(_ context.Context, sd *gRPC.OnHealthCheck) (*gRPC.Response, error) {
	success := getCollector().UpdateInstanceTimestamp(sd.Instance)
	return &gRPC.Response{Success: success}, nil
}

// implement the gRPC.EventClient interface
func (ev *EventServer) GetInstances(_ context.Context, rn *gRPC.RoleData) (*gRPC.Instances, error) {
	role := rn.GetRole()
	insts := getCollector().GetInstances(role)

	rItems := make([]*gRPC.Instance, 0, len(insts))
	for _, inst := range insts {
		rItems = append(rItems, &gRPC.Instance{
			Role:      inst.Role,
			Id:        inst.ID,
			Ip:        inst.Ip,
			Cluster:   inst.Cluster,
			CreatedAt: timestamppb.New(inst.CreatedAt),
		})
	}

	return &gRPC.Instances{Items: rItems}, nil
}
