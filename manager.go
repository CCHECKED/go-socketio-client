package socketio

import (
	"crypto/tls"
	"encoding/json"
	"github.com/CCHECKED/go-socketio-client/consts"
	"github.com/CCHECKED/go-socketio-client/logger"
	"net/http"
	"net/url"
	"time"
)

type SocketManager struct {
	config                  *SocketManagerConfig
	socketURL               *url.URL
	nsps                    map[string]*SocketIoClient
	client                  *SocketIoClient
	engine                  *SocketEngine
	status                  consts.SocketStatus
	currentReconnectAttempt int
	logger                  *logger.Logger
}

type SocketManagerConfig struct {
	Reconnects    bool
	ReconnectWait time.Duration
	ConnectParams map[string]string
	AllowUpgrade  bool
	Tls           *tls.Config
	Headers       *http.Header
	Debug         bool
}

func NewSocketManager(socketURL *url.URL, config *SocketManagerConfig) *SocketManager {
	if socketURL.Scheme != "http" && socketURL.Scheme != "https" {
		panic("socketURL.Scheme must be http or https")
	}

	if socketURL.Path == "" {
		panic("socketURL.Path must not be empty")
	}

	if &config.AllowUpgrade == nil {
		config.AllowUpgrade = true
	}

	if &config.Debug == nil {
		config.Debug = false
	}

	// Setting default params
	if &config.Reconnects == nil {
		config.Reconnects = true
	}

	if config.ReconnectWait == time.Duration(0) {
		config.ReconnectWait = 5 * time.Second
	}

	loggerLever := logger.LevelError
	if config.Debug {
		loggerLever = logger.LevelDebug
	}

	managerLogger := logger.NewLogger(loggerLever)

	// Create new instance
	self := &SocketManager{
		config:                  config,
		socketURL:               socketURL,
		nsps:                    make(map[string]*SocketIoClient),
		currentReconnectAttempt: 0,
		logger:                  managerLogger,
		status:                  consts.NOT_CONNECTED,
	}

	return self
}

func (manager *SocketManager) DefaultSocket(forNamespace ...string) *SocketIoClient {
	namespace := "/"
	if len(forNamespace) > 0 {
		namespace = forNamespace[0]
	}
	return manager.socket(namespace)
}

func (manager *SocketManager) socket(namespace string) *SocketIoClient {
	// TODO: Don't use panic
	if len(namespace) == 0 || namespace[0] != '/' {
		// TODO: change panic on log error
		panic("forNamespace must have a leading /")
	}

	socket, exitst := manager.nsps[namespace]
	if exitst {
		return socket
	}

	manager.connect()

	client := NewSocketIoClient(manager, namespace)

	manager.nsps[namespace] = client

	return client
}

func (manager *SocketManager) addEngine() {
	//fmt.Println("SocketManager addEngine")
	manager.logger.Debug(manager._name(), "addEngine")
	manager.engine = NewSocketIoEngine(manager, manager.socketURL, manager.config)
}

func (manager *SocketManager) connect() {
	manager.logger.Debug(manager._name(), "connect")
	if manager.status == consts.CONNECTED || (manager.status == consts.CONNECTING && manager.currentReconnectAttempt == 0) {
		return
	}

	if manager.engine == nil {
		manager.addEngine()
	}

	manager.status = consts.CONNECTING

	for {
		err := manager.engine.connect()
		if err != nil {
			manager.logger.Error(manager._name(), "Failed polling: "+err.Error())
			manager.logger.Error(manager._name(), "Reconnect after", manager.config.ReconnectWait.String()+"...")
			time.Sleep(manager.config.ReconnectWait)
			continue
		}
		break
	}
}

func (manager *SocketManager) connectSocket(socket *SocketIoClient, withPayload map[string]any) {
	manager.logger.Debug(manager._name(), "connectSocket")

	if manager.status != consts.CONNECTED {
		manager.connect()
	}

	payload := ""
	if withPayload != nil {
		payloadBytes, _ := json.Marshal(&withPayload)
		payload = string(payloadBytes)
	}
	// TODO: Add error handle
	manager.engine.SendConnect(socket.namespace, payload)
}

func (manager *SocketManager) disconnect() {
	manager.logger.Debug(manager._name(), "disconnect")
	manager.status = consts.DISCONNECTED

	if manager.engine != nil {
		manager.engine.disconnect("Disconnect")
	}
}

func (manager *SocketManager) EngineDidError(reason string) {}
func (manager *SocketManager) EngineDidClose(reason string) {
	manager.logger.Debug(manager._name(), "EngineDidClose")
	manager.status = consts.NOT_CONNECTED
	for _, client := range manager.nsps {
		client.disconnect()
	}
}
func (manager *SocketManager) EngineDidOpen(reason string) {
	manager.logger.Debug(manager._name(), "EngineDidOpen")

	manager.status = consts.CONNECTED

	for _, client := range manager.nsps {
		client.connect()
	}
}
func (manager *SocketManager) EngineDidReceivePing() {
	manager.logger.Debug(manager._name(), "EngineDidReceivePing")
	for _, client := range manager.nsps {
		if handlers, exists := client.handlersService[consts.PING]; exists {
			for _, handler := range handlers {
				handler()
			}
		}
	}
}
func (manager *SocketManager) EngineDidReceivePong() {
	manager.logger.Debug(manager._name(), "EngineDidReceivePong")
	for _, client := range manager.nsps {
		if handlers, exists := client.handlersService[consts.PONG]; exists {
			for _, handler := range handlers {
				handler()
			}
		}
	}
}
func (manager *SocketManager) EngineDidSendPing()                                  {}
func (manager *SocketManager) EngineDidSendPong()                                  {}
func (manager *SocketManager) ParseEngineMessage(msg string)                       {}
func (manager *SocketManager) EngineDidWebsocketUpgrade(headers map[string]string) {}
func (manager *SocketManager) EngineDidReceiveMessageEvent(namespace string, event string, data interface{}) {
	manager.logger.Debug(manager._name(), "EngineDidReceiveMessageEvent")
	if client, ok := manager.nsps[namespace]; ok {
		if handlers, exists := client.handlers[event]; exists {
			for _, handler := range handlers {
				handler(data)
			}
		}
	}
}

func (manager *SocketManager) EngineDidMessageConnect(namespace string, message string) {
	if client, ok := manager.nsps[namespace]; ok {
		client.setStatusConnect(consts.CONNECTED)
	}
}

func (manager *SocketManager) _name() string {
	return "SocketManager"
}
