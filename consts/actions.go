package consts

type SocketActionType rune

const (
	ACTION_CONNECT      SocketActionType = '0'
	ACTION_DISCONNECT                    = '1'
	ACTION_EVENT                         = '2'
	ACTION_ACK                           = '3'
	ACTION_ERROR                         = '4'
	ACTION_BINARY_EVENT                  = '5'
	ACTION_BINARY_ACK                    = '6'
)

func (t SocketActionType) Byte() []byte {
	return []byte{byte(t)}
}
