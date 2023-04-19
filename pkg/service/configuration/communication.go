package configuration

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Role int

const (
	Indexer Role = iota
	MemPool
	Proxy
	Subscriber
	WsSubscriber
	KeyManager
	WebSocket
	AirDropper

	Unknown
)

var (
	allRoles []Role = []Role{Indexer, MemPool, Proxy, Subscriber, WsSubscriber, KeyManager, WebSocket, AirDropper}
)

func (r Role) String() string {
	switch r {
	case Indexer:
		return "indexer"
	case MemPool:
		return "mempool"
	case Proxy:
		return "proxy"
	case Subscriber:
		return "subscriber"
	case WsSubscriber:
		return "wssubscriber"
	case KeyManager:
		return "keymanager"
	case WebSocket:
		return "websocket"
	case AirDropper:
		return "airdropper"
	}
	return "unknown"
}

func FromString(name string) Role {
	switch name {
	case "indexer":
		return Indexer
	case "mempool":
		return MemPool
	case "proxy":
		return Proxy
	case "subscriber":
		return Subscriber
	case "wssubscriber":
		return WsSubscriber
	case "keymanager":
		return KeyManager
	case "websocket":
		return WebSocket
	case "airdropper":
		return AirDropper
	}
	return Unknown
}

const (
	defaultCurRoleServiceCount         = "1"
	defaultCurRoleIndex                = "0"
	defaultCommunicationServerAddress  = "127.0.0.1"
	defaultCommunicationServerPort     = "10200"
	defaultCommunicationEndpointServer = "127.0.0.1"
	defaultCommunicationEndpointPort   = "1055"
)

// COMMUNICATION PROTOCOL CONFIG
type ProtocolConfiguration struct {
	Role       Role
	Ip         string
	InstanceID string
	Cluster    string
	CreatedAt  time.Time
}

// LOAD COMMUNICATION  PROTOCOL CONFIGURATION
func (c *ProtocolConfiguration) LoadProtocolConfig(name string, index int) (err error) {
	name = strings.ToUpper(name)
	c.CreatedAt = time.Now().UTC()
	c.Ip = os.Getenv(fmt.Sprintf("NS_COMMUNICATION_%s_%d_IP", name, index))
	c.InstanceID = os.Getenv(fmt.Sprintf("NS_COMMUNICATION_%s_%d_INSTANCE_ID", name, index))
	c.Cluster = os.Getenv(fmt.Sprintf("NS_COMMUNICATION_%s_%d_CLUSTER_NAME", name, index))
	if c.InstanceID == "" {
		c.InstanceID = generateID()
	}
	return nil
}

type CommunicationProtocolConfiguration struct {
	CommunicationServerPort         string
	CommunicationEndpointServerPort string
	MainConfig                      *ProtocolConfiguration
	RelativeConfigs                 map[Role][]ProtocolConfiguration
}

// LOAD PROTOCOL CONFIGURATION
func (c *CommunicationProtocolConfiguration) LoadCommunicationProtocolConfig(name string) (err error) {
	prType := FromString(name)
	for _, role := range allRoles {
		prName := role.String()
		sName := strings.ToUpper(prName)
		prCount, found := os.LookupEnv(fmt.Sprintf("NS_COMMUNICATION_%s_COUNT", sName))
		if !found && role == prType {
			prCount = defaultCurRoleServiceCount
		}
		count, err := strconv.Atoi(prCount)
		if err != nil {
			continue // continue read other roles
		}
		for index := 0; index < count; index++ {
			relCfg := ProtocolConfiguration{}
			relCfg.LoadProtocolConfig(prName, index)
			c.RelativeConfigs[role] = append(c.RelativeConfigs[role], relCfg)
		}
	}

	name = strings.ToUpper(name)
	prCurIndex, found := os.LookupEnv(fmt.Sprintf("NS_COMMUNICATION_CUR_%s_INDEX", name))
	if !found {
		prCurIndex = defaultCurRoleIndex
	}
	index, err := strconv.Atoi(prCurIndex)
	if err != nil {
		return errors.New("config parse error: current index is not a number")
	}
	if index >= len(c.RelativeConfigs[prType]) {
		return errors.New("config parse error: current index is not less than roles count")
	}

	c.MainConfig = &c.RelativeConfigs[prType][index]

	sAddress := os.Getenv("NS_COMMUNICATION_SERVER_LISTEN_ADDRESS")
	if sAddress == "" {
		sAddress = defaultCommunicationServerAddress
	}
	sPort := os.Getenv("NS_COMMUNICATION_SERVER_LISTEN_PORT")
	if sPort == "" {
		sPort = defaultCommunicationServerPort
	}
	port, err := strconv.Atoi(sPort)
	if err != nil {
		port = 0
	}
	c.CommunicationServerPort = fmt.Sprintf("%s:%d", sAddress, port)

	cesAddress := os.Getenv("NS_COMMUNICATION_ENDPOINT_SERVER_LISTEN_ADDRESS")
	if cesAddress == "" {
		cesAddress = defaultCommunicationEndpointServer
	}
	cesPort := os.Getenv("NS_COMMUNICATION_ENDPOINT_SERVER_LISTEN_PORT")
	if cesPort == "" {
		cesPort = defaultCommunicationEndpointPort
	}
	eport, err := strconv.Atoi(cesPort)
	if err != nil {
		eport = 0
	}
	c.CommunicationEndpointServerPort = fmt.Sprintf("%s:%d", cesAddress, eport)

	return nil
}

// LOAD COMMUNICATION PROTOCOL CONFIGURATION
func (c *ServiceConfiguration) loadCommunicationProtocolConfiguration(serviceName string) error {
	return c.CommunicationProtocol.LoadCommunicationProtocolConfig(serviceName)
}

var (
	instanceIDDefaultLenght = 24
	allrunelist             = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
)

func generateID() string {
	b := make([]rune, instanceIDDefaultLenght)
	for i := range b {
		b[i] = allrunelist[rand.Intn(len(allrunelist))]
	}
	return string(b)
}
