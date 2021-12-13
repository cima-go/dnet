package dnet

import (
	"context"

	"github.com/cima-go/dnet/spec"
)

type Peer struct {
	finder   *Discovery
	address  string
	identify *spec.Identify
}

func (p *Peer) Address() string {
	return p.address
}

func (p *Peer) Identify() *spec.Identify {
	return p.identify
}

func (p *Peer) Request(ctx context.Context, name string, req []byte) (resp []byte, err error) {
	return p.finder.Unicast(ctx, p.identify, name, req)
}
