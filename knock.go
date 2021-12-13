package dnet

import (
	"github.com/golang/protobuf/proto"

	"github.com/cima-go/dnet/lan"
	"github.com/cima-go/dnet/spec"
)

func knockRequest() []byte {
	return encodePacket(KNOCKING, nil)
}

func knockResponse(peer lan.Discovered, identify *spec.Identify) {
	if bs, err := proto.Marshal(identify); err == nil {
		peer.Reply(encodePacket(IDENTIFY, bs))
	}
}
