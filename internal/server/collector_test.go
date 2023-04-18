package server

import (
	"context"
	"testing"
	"time"

	"github.com/neonlabsorg/neon-proxy/pkg/gRPC"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toGRPCInstance(inst Instance) *gRPC.Instance {
	return &gRPC.Instance{
		Role:      inst.Role,
		Ip:        inst.Ip,
		Id:        inst.ID,
		Cluster:   inst.Cluster,
		CreatedAt: timestamppb.New(inst.CreatedAt),
	}
}

func TestCollectorAddRemoveInstance(t *testing.T) {
	now := time.Now().UTC()
	inst := Instance{
		Role:      1,
		Ip:        "localhost:8080",
		ID:        "3956f893fc2800400e2a2bf847f548b9679ac50e",
		Cluster:   "CL1",
		CreatedAt: now,
	}

	success := getCollector().AddInstance(toGRPCInstance(inst))

	// test GetAllInstances
	instances := getCollector().GetAllInstances()
	assert.True(t, success)
	assert.Len(t, instances, 1)

	instance := instances[0]

	assert.Equal(t, int32(1), instance.Role)
	assert.Equal(t, "localhost:8080", instance.Ip)
	assert.Equal(t, "3956f893fc2800400e2a2bf847f548b9679ac50e", instance.ID)
	assert.Equal(t, "CL1", instance.Cluster)
	assert.Equal(t, now, instance.CreatedAt)

	// test GetInstances
	instances = getCollector().GetInstances(1)
	assert.Len(t, instances, 1)

	instance = instances[0]

	assert.Equal(t, int32(1), instance.Role)
	assert.Equal(t, "localhost:8080", instance.Ip)
	assert.Equal(t, "3956f893fc2800400e2a2bf847f548b9679ac50e", instance.ID)
	assert.Equal(t, "CL1", instance.Cluster)
	assert.Equal(t, now, instance.CreatedAt)

	// removeInstance
	instance.CreatedAt = now.Add(10 * time.Millisecond)
	success = getCollector().RemoveInstance(toGRPCInstance(instance))

	instances = getCollector().GetAllInstances()
	assert.True(t, success)
	assert.Empty(t, instances)

	instances = getCollector().GetInstances(1)
	assert.Empty(t, instances)
}

func TestEventServerCreateShutDown(t *testing.T) {
	ev := EventServer{}
	now := time.Now().UTC()
	gRPCInstance := &gRPC.Instance{
		Role:      2,
		Ip:        "localhost:5050",
		Id:        "f847f548b9679ac50e3956f893fc2800400e2a2b",
		Cluster:   "CL2",
		CreatedAt: timestamppb.Now(),
	}

	// test Create
	onCreate := &gRPC.OnCreate{Instance: gRPCInstance}

	resp, err := ev.AfterCreate(context.TODO(), onCreate)
	assert.NoError(t, err)
	assert.True(t, resp.Success)

	// test Get Instances
	instances, err := ev.GetInstances(context.TODO(), &gRPC.RoleData{Role: 2})
	assert.NoError(t, err)
	assert.Len(t, instances.Items, 1)

	inst := instances.Items[0]

	assert.Equal(t, int32(2), inst.Role)
	assert.Equal(t, "localhost:5050", inst.Ip)
	assert.Equal(t, "f847f548b9679ac50e3956f893fc2800400e2a2b", inst.Id)
	assert.Equal(t, "CL2", inst.Cluster)
	assert.WithinDuration(t, inst.CreatedAt.AsTime(), now, 1*time.Millisecond)

	// Test Shut Down
	onShutDown := &gRPC.OnShutDown{Instance: gRPCInstance}
	resp, err = ev.BeforeShutDown(context.TODO(), onShutDown)
	assert.NoError(t, err)
	assert.True(t, resp.Success)

	// test Get Instances
	instances, err = ev.GetInstances(context.TODO(), &gRPC.RoleData{Role: 2})
	assert.NoError(t, err)
	assert.Empty(t, instances.Items)
}

func TestEventServerHealthCheck(t *testing.T) {
	ev := EventServer{}
	gRPCInst := &gRPC.Instance{
		Role:      3,
		Ip:        "localhost:0001",
		Id:        "79ac50e3956f893fc2800400e2a2bf847f548b96",
		Cluster:   "CL3",
		CreatedAt: timestamppb.Now(),
	}

	// test Create
	onCreate := &gRPC.OnCreate{Instance: gRPCInst}

	resp, err := ev.AfterCreate(context.TODO(), onCreate)
	assert.NoError(t, err)
	assert.True(t, resp.Success)

	// test Get Instances 1 element
	roles, err := ev.GetInstances(context.TODO(), &gRPC.RoleData{Role: 3})
	assert.NoError(t, err)
	assert.Len(t, roles.Items, 1)

	time.Sleep(50 * time.Millisecond)

	resp, err = ev.HealthCheck(context.TODO(), &gRPC.OnHealthCheck{Instance: gRPCInst})
	assert.NoError(t, err)
	assert.True(t, resp.Success)

	// test Get Instances 1 element
	roles, err = ev.GetInstances(context.TODO(), &gRPC.RoleData{Role: 3})
	assert.NoError(t, err)
	assert.Len(t, roles.Items, 1)

	time.Sleep(3*time.Second + 10*time.Millisecond) // sorry for sleeping unit-tests
	resp, err = ev.HealthCheck(context.TODO(), &gRPC.OnHealthCheck{Instance: gRPCInst})
	assert.NoError(t, err)
	assert.False(t, resp.Success)

	// test Get Instances empty list
	roles, err = ev.GetInstances(context.TODO(), &gRPC.RoleData{Role: 3})
	assert.NoError(t, err)
	assert.Empty(t, roles.Items)
}
