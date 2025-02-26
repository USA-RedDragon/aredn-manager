package arednlink

import "net"

type Command byte

func (c Command) String() string {
	return string(c)
}

const (
	CommandVersion        Command = 'V'
	CommandUpdateHosts    Command = 'H'
	CommandUpdateServices Command = 'U'
	CommandSync           Command = 'S'
	CommandKeepAlive      Command = 'K'
)

type Message struct {
	Length  uint16
	Command Command
	Payload []byte
	Source  net.IP
	Hops    uint8
	ConnID  string // Internal use for tracking connections
}

func (m *Message) Bytes() []byte {
	buf := make([]byte, 8)
	buf[0] = byte(m.Length >> 8)
	buf[1] = byte(m.Length)
	buf[2] = byte(m.Command)
	buf[3] = byte(m.Hops)
	copy(buf[4:8], m.Source)
	return append(buf, m.Payload...)
}
