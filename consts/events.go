package consts

type ClientServiceEvents int

const (
	OPEN ClientServiceEvents = iota
	CLOSE
	PING
	PONG
	MESSAGE
	UPGRADE
	NOOP
)
