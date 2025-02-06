package socketio

import (
	"github.com/CCHECKED/go-socketio-client/consts"
)

var loggerName = "SocketIoClient"

type SocketIoClient struct {
	handlers        map[string][]func(data interface{})
	handlersService map[consts.ClientServiceEvents][]func()
	manager         *SocketManager
	namespace       string
	status          consts.SocketStatus
}

func NewSocketIoClient(manager *SocketManager, namespace string /*, cb func()*/) *SocketIoClient {
	self := &SocketIoClient{
		manager:         manager,
		handlers:        make(map[string][]func(interface{})),
		handlersService: make(map[consts.ClientServiceEvents][]func()),
		namespace:       namespace,
	}
	manager.logger.Debug(self._name(), "NewSocketIoClient")
	self.connect()
	return self
}

func (client *SocketIoClient) On(event string, handler func(interface{})) {
	client.handlers[event] = append(client.handlers[event], handler)
}

func (client *SocketIoClient) OnService(event consts.ClientServiceEvents, handler func()) {
	client.handlersService[event] = append(client.handlersService[event], handler)
}

func (client *SocketIoClient) Emit(event string, data interface{}) {
	//
}

func (client *SocketIoClient) setStatusConnect(status consts.SocketStatus) {
	client.manager.logger.Debug(client._name(), "setStatusConnect namespace:", client.namespace, ", status:", status)
	client.status = status
}

func (client *SocketIoClient) connect() {
	client.manager.logger.Debug(client._name(), "connect")
	if client.manager == nil || client.status == consts.CONNECTED {
		return
	}

	client.setStatusConnect(consts.CONNECTING)
	client.joinNamespace(nil)
}

func (client *SocketIoClient) disconnect() {
	client.status = consts.DISCONNECTED
}

func (client *SocketIoClient) joinNamespace(withPayload map[string]any) {
	client.manager.logger.Debug(client._name(), "joinNamespace:", client.namespace, "Payload:", withPayload)
	client.manager.connectSocket(client, withPayload)
}

func (client *SocketIoClient) _name() string {
	return "SocketIoClient"
}
