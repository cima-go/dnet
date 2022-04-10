package dnet

import (
	"bytes"
	"errors"

	"github.com/golang/protobuf/proto"

	"github.com/cima-go/dnet/lan"
	"github.com/cima-go/dnet/spec"
)

var magic = "  DNETv1"

// CType is command type
type CType uint8

const (
	UNKNOWN  CType = iota // default
	KNOCKING              // Client knocking
	IDENTIFY              // Server identify
	REQUEST               // RPC Request
	REPLY                 // RPC Response
)

func knockRequest() []byte {
	return encodePacket(KNOCKING, nil)
}

func knockResponse(peer lan.Discovered, identify *spec.Identify) {
	if bs, err := proto.Marshal(identify); err == nil {
		peer.Reply(encodePacket(IDENTIFY, bs))
	}
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
