package dnet

import (
	"bytes"
	"errors"
)

var magic = "  CIMA"

// CType is command type
type CType uint8

const (
	UNKNOWN  CType = iota // default
	KNOCKING              // Client knocking
	IDENTIFY              // Server identify
	REQUEST               // RPC Request
	REPLY                 // RPC Response
)

type Options struct {
	Addr string
	Port string
}

func encodePacket(cmd CType, data []byte) []byte {
	buf := bytes.NewBufferString(magic)
	buf.WriteByte(byte(cmd))
	buf.Write(data)
	return buf.Bytes()
}

func decodePacket(payload []byte) (cmd CType, data []byte, err error) {
	if !bytes.HasPrefix(payload, []byte(magic)) || len(payload) <= len(magic) {
		err = errors.New("invalid packet")
		return
	}

	cmd = CType(payload[len(magic)])
	data = payload[len(magic)+1:]

	return
}
