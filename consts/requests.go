package consts

type SocketRequestType rune

const (
	REQUEST_OPEN    SocketRequestType = '0'
	REQUEST_CLOSE                     = '1'
	REQUEST_PING                      = '2'
	REQUEST_PONG                      = '3'
	REQUEST_MESSAGE                   = '4'
	REQUEST_UPGRADE                   = '5'
	REQUEST_NOOP                      = '6'
)

func (t SocketRequestType) Byte() []byte {
	return []byte{byte(t)}
}
