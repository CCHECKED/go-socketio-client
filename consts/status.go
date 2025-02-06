package consts

type SocketStatus int

const (
	NOT_CONNECTED SocketStatus = iota
	DISCONNECTED
	CONNECTING
	CONNECTED
)
