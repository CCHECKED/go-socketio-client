package socketio

//TODO: менеджер должен
//1. Заниматься подключением!
//2. Заниматься переподключением
//3. Следить за статусом подключения
//
//А НЕ ДВИЖОК

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/CCHECKED/socketio-client/consts"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type SocketEngine struct {
	closed       bool
	connected    bool
	polling      bool
	probing      bool
	sid          string
	socketPath   string
	urlPolling   url.URL
	urlWebSocket url.URL
	pingInterval time.Duration
	pingTimeout  time.Duration
	ws           *websocket.Conn
	manager      *SocketManager
	url          *url.URL
	config       *SocketManagerConfig
	httpClient   *http.Client
	mutex        sync.Mutex
}

func NewSocketIoEngine(client *SocketManager, url *url.URL, config *SocketManagerConfig) *SocketEngine {
	if config.ConnectParams == nil {
		config.ConnectParams = make(map[string]string)
	}

	config.ConnectParams["EIO"] = "4"

	httpTransport := &http.Transport{
		TLSClientConfig: config.Tls,
	}

	httpClient := &http.Client{Transport: httpTransport}

	self := &SocketEngine{
		manager:    client,
		url:        url,
		config:     config,
		httpClient: httpClient,
		polling:    true,
		probing:    false,
		connected:  false,
		closed:     false,
	}

	self.createURLs()

	return self
}

func (engine *SocketEngine) connect() error {
	engine.manager.logger.Debug(engine._name(), "connect")
	err := engine._polling()
	if err != nil {
		return err
	}

	go engine.listen()
	return nil
}

func (engine *SocketEngine) _polling() error {
	req, err := http.NewRequest("GET", engine.urlPolling.String(), nil)
	if err != nil {
		return err
		//fmt.Println("Error creating polling request: " + err.Error())
		//return
	}

	resp, err := engine.httpClient.Do(req)
	if err != nil {
		time.Sleep(time.Second)
		engine.manager.EngineDidClose("Error polling request")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 400 {
		engine.createURLs()
	}

	if resp.StatusCode != 200 {
		time.Sleep(time.Second)
		return errors.New("polling failed, status code: " + resp.Status)
	}

	bodyResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		//fmt.Println("Error reading response body: " + err.Error())
		return err
	}

	engine.parseEngineMessage(string(bodyResponse))
	return nil
}

func (engine *SocketEngine) pollingSendData(message string) {
	engine.manager.logger.Debug(engine._name(), "pollingSendData:", message)
	req, err := http.NewRequest("POST", engine.urlPolling.String(), strings.NewReader(message))
	if err != nil {
		panic("Error creating polling request: " + err.Error())
	}

	resp, err := engine.httpClient.Do(req)
	if err != nil {
		panic("Error polling request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		panic("Error polling status: " + resp.Status)
	}

	bodyResponse, err := io.ReadAll(resp.Body)
	if err != nil {

		panic("Error reading response body: " + err.Error())
	}

	if string(bodyResponse) == "ok" {

	}
}

func (engine *SocketEngine) createWebSocketAndConnect() {
	dialer := websocket.Dialer{
		TLSClientConfig: engine.config.Tls,
		Proxy:           http.ProxyFromEnvironment,
	}

	conn, _, err := dialer.Dial(engine.urlWebSocket.String(), *engine.config.Headers)
	if err != nil {
		engine.didError("Error connecting to WebSocket")
		fmt.Println("Error connecting to WebSocket: " + err.Error())
		return
	}

	engine.mutex.Lock()
	engine.ws = conn
	engine.mutex.Unlock()

	engine.manager.logger.Debug(engine._name(), "createWebSocketAndConnect ")
	engine.manager.logger.Info(engine._name(), "Upgraded to WebSocket")
}

func (engine *SocketEngine) _probing() {
	data := append([]byte{consts.REQUEST_PING}, []byte("probe")...)
	err := engine.ws.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		fmt.Println("Error probing websocket: " + err.Error())
		return
	}

	engine.mutex.Lock()
	engine.probing = true
	engine.mutex.Unlock()
	engine.manager.logger.Debug(engine._name(), "Probing successfuly")
}

func (engine *SocketEngine) _upgrade() {
	err := engine.ws.WriteMessage(websocket.TextMessage, consts.SocketRequestType(consts.REQUEST_UPGRADE).Byte())
	if err != nil {
		fmt.Println("Error upgrading websocket: " + err.Error())
	}
}

func (engine *SocketEngine) ping() {
	if engine.config.AllowUpgrade && engine.ws != nil {
		err := engine.ws.WriteMessage(websocket.TextMessage, []byte{consts.REQUEST_PING})
		if err != nil {
			fmt.Println("Error ping websocket: " + err.Error())
		}
	} else {
		engine.pollingSendData(string([]byte{consts.REQUEST_PING}))
	}
}

func (engine *SocketEngine) pong() {
	if engine.config.AllowUpgrade && engine.ws != nil {
		err := engine.ws.WriteMessage(websocket.TextMessage, []byte{consts.REQUEST_PONG})
		if err != nil {
			fmt.Println("Error pong websocket: " + err.Error())
		}
	} else {
		engine.pollingSendData(string([]byte{consts.REQUEST_PONG}))
	}
}

func (engine *SocketEngine) listen() {
	for {
		select {
		default:
			if engine.closed {
				return
			}
			if engine.config.AllowUpgrade && engine.ws != nil {
				_, message, err := engine.ws.ReadMessage()
				if err != nil {
					engine.manager.logger.Error(engine._name(), "Error reading from websocket: "+err.Error())
					engine.mutex.Lock()
					engine.ws = nil
					engine.connected = false
					engine.mutex.Unlock()
					engine.manager.EngineDidClose("Connection closed")

					// TODO: Сообщаем всем хандлерам что соединение оборвано
					continue
				}
				engine.parseEngineMessage(string(message))
			} else {
				err := engine._polling()
				if err != nil {
					engine.manager.logger.Error(engine._name(), err)
				}
			}
		}
	}
}

func (engine *SocketEngine) disconnect(reason string) {
	engine.mutex.Lock()
	engine.closed = true
	err := engine.ws.Close()
	if err != nil {
	}
	engine.ws = nil
	engine.connected = false
	engine.sid = ""
	engine.probing = false
	engine.polling = false
	engine.mutex.Unlock()

	engine.manager.EngineDidClose(reason)
}

func (engine *SocketEngine) parseEngineMessage(msg string) {
	engine.manager.logger.Debug(engine._name(), "parseEngineMessage:", msg)
	switch msg[0] {
	case uint8(consts.REQUEST_OPEN):
		engine.handleOpen(msg[1:])
	case consts.REQUEST_CLOSE:
		engine.handleClose(msg)
	case consts.REQUEST_PING:
		engine.handlePing(msg)
	case consts.REQUEST_PONG:
		engine.handlePong(msg[1:])
	case consts.REQUEST_MESSAGE:
		switch msg[1] {
		case uint8(consts.ACTION_CONNECT):
			namespace, message := engine.parseNamespaceAndMessage(msg, 2)
			engine.handleMessageConnect(namespace, message)
		case consts.ACTION_EVENT:
			namespace, message := engine.parseNamespaceAndMessage(msg, 2)
			engine.handleMessageEvent(namespace, message)
		case consts.ACTION_BINARY_EVENT:
			engine.handleMessageBinaryEvent(msg[2:])
		}
	}
}

func (engine *SocketEngine) parseNamespaceAndMessage(message string, indexStart int) (string, string) {
	indexMessage := indexStart
	namespace := "/"
	if message[indexStart] == '/' {
		index := strings.IndexRune(message, ',')
		if index == -1 {
			engine.manager.logger.Error(engine._name(), "parseEngineMessage Error parsing event")
		}
		namespace = message[2:index]
		indexMessage = index + 1
	}

	return namespace, message[indexMessage:]
}

func (engine *SocketEngine) createURLs() {
	urlPolling := *engine.url
	urlWebSocket := *engine.url

	if engine.url.Scheme == "https" {
		urlPolling.Scheme = "https"
		urlWebSocket.Scheme = "wss"
	} else {
		urlPolling.Scheme = "http"
		urlWebSocket.Scheme = "ws"
	}

	queryParams := url.Values{}
	if engine.config.ConnectParams != nil {
		for key, value := range engine.config.ConnectParams {
			queryParams.Set(key, value)
		}
	}

	queryParams.Set("transport", "websocket")
	urlWebSocket.RawQuery = queryParams.Encode()
	queryParams.Set("transport", "polling")
	urlPolling.RawQuery = queryParams.Encode()

	engine.mutex.Lock()
	engine.urlPolling = urlPolling
	engine.urlWebSocket = urlWebSocket
	engine.mutex.Unlock()
	engine.manager.logger.Debug(engine._name(), "URLs successfully created")
}

func (engine *SocketEngine) patchUrls() {
	if engine.sid == "" {
		panic("Sid is empty")
	}
	queryParamsPolling := engine.urlPolling.Query()
	queryParamsPolling.Set("sid", engine.sid)

	queryParamsWebsocket := engine.urlWebSocket.Query()
	queryParamsWebsocket.Set("sid", engine.sid)

	engine.mutex.Lock()
	engine.urlPolling.RawQuery = queryParamsPolling.Encode()
	engine.urlWebSocket.RawQuery = queryParamsWebsocket.Encode()
	engine.mutex.Unlock()
}

func (engine *SocketEngine) handleOpen(msg string) {
	var result struct {
		Sid          string   `json:"sid"`
		Upgrades     []string `json:"upgrades"`
		PingInterval uint64   `json:"pingInterval"`
		PingTimeout  uint64   `json:"pingTimeout"`
	}

	if err := json.Unmarshal([]byte(msg), &result); err != nil {
		engine.didError("Error parsing open packet")
		return
	}

	engine.mutex.Lock()
	engine.pingInterval = time.Duration(result.PingInterval * 1000000)
	engine.pingTimeout = time.Duration(result.PingTimeout * 1000000)
	engine.sid = result.Sid
	engine.manager.logger.Info(engine._name(), "Setting new sid:", engine.sid)
	engine.mutex.Unlock()
	engine.patchUrls()

	if Contains(result.Upgrades, "websocket") && engine.config.AllowUpgrade {
		engine.manager.logger.Info(engine._name(), "Using WebSocket")
		engine.createWebSocketAndConnect()
		engine._probing()
		engine._upgrade()
	} else {
		engine.manager.logger.Info(engine._name(), "Using Polling")
	}

	engine.manager.logger.Debug(engine._name(), "Open packet successfully")
	engine.manager.EngineDidOpen("Connect")
}

func (engine *SocketEngine) handleClose(msg string) {
	engine.mutex.Lock()
	engine.ws = nil
	engine.connected = false
	engine.sid = ""
	engine.mutex.Unlock()
}
func (engine *SocketEngine) handlePing(msg string) {
	engine.pong()
	engine.manager.EngineDidReceivePing()
}
func (engine *SocketEngine) handlePong(msg string) {
	if msg == "probe" {
		engine.manager.logger.Debug("SocketEngine", "Probe packet successfully")
		engine._upgrade()
		return
	}
	engine.manager.EngineDidReceivePong()
}
func (engine *SocketEngine) handleMessageConnect(namespace string, message string) {
	engine.manager.EngineDidMessageConnect(namespace, message)
}

func (engine *SocketEngine) handleMessageEvent(namespace string, msg string) {
	engine.manager.logger.Debug(engine._name(), "handleMessageEvent Namespace:", namespace, "Message:", msg)
	var eventData []interface{}
	if err := json.Unmarshal([]byte(msg), &eventData); err != nil {
		engine.manager.logger.Error(engine._name(), "Failed to unmarshal event data")
		return
	}
	engine.manager.EngineDidReceiveMessageEvent(namespace, eventData[0].(string), eventData[1])
}
func (engine *SocketEngine) handleMessageBinaryEvent(msg string) {}

func (engine *SocketEngine) didError(reason string) {
	engine.manager.EngineDidError(reason)
	engine.disconnect(reason)
}

func (engine *SocketEngine) SendConnect(namespace string, data string) {

	engine.manager.logger.Debug(engine._name(), "SendConnect")
	packetConnect := append(
		append(
			consts.SocketRequestType(consts.REQUEST_MESSAGE).Byte(),
			consts.SocketActionType(consts.ACTION_CONNECT).Byte()...,
		),
	)

	if namespace != "/" {
		packetConnect = append(
			packetConnect,
			append([]byte(namespace))...,
		)
	}

	if data != "" {
		packetConnect = append(
			packetConnect,
			append([]byte(","), []byte(data)...)...,
		)
	}

	engine.connected = true
	if engine.ws != nil && engine.connected {
		engine.manager.logger.Debug(engine._name(), "SendConnect over WebSocket:", string(packetConnect))
		err := engine.ws.WriteMessage(websocket.TextMessage, packetConnect)
		if err != nil {
			engine.manager.logger.Error(engine._name(), "Error probing websocket: ", err.Error())
		}
	} else {
		engine.manager.logger.Debug(engine._name(), "SendConnect polling:", string(packetConnect))
		engine.pollingSendData(string(packetConnect))
		engine.connected = true
	}
}

func (engine *SocketEngine) _name() string {
	return "SocketEngine"
}
