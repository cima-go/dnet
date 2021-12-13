package dnet

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"go.uber.org/zap"

	"github.com/cima-go/dnet/lan"
	"github.com/cima-go/dnet/rpc"
	"github.com/cima-go/dnet/spec"
)

func Finder(opts Options, logger *zap.Logger) *Discovery {
	return &Discovery{opts, logger.With(zap.String("mod", "discovery"))}
}

type Discovery struct {
	opts   Options
	logger *zap.Logger
}

func (s *Discovery) Peers(timeout time.Duration, notify func(*Peer)) ([]*Peer, error) {
	var peers []*Peer
	var found = make(map[string]struct{})

	if _, err := lan.Discover(lan.Settings{
		MulticastAddress: s.opts.Addr,
		Port:             s.opts.Port,
		TimeLimit:        timeout,
		Payload:          knockRequest(),
		Notify: func(peer lan.Discovered) {
			cmd, data, err := decodePacket(peer.Payload)
			if err != nil {
				s.logger.With(zap.Error(err)).Debug("invalid packet")
				return
			}

			if cmd == IDENTIFY {
				identify := &spec.Identify{}
				if err := proto.Unmarshal(data, identify); err != nil {
					s.logger.With(zap.Error(err)).Info("message decode failed")
					return
				}

				if _, exists := found[string(identify.MachineID)]; !exists {
					s.logger.With(zap.String("addr", peer.Address)).Debug("found new peer")
					found[string(identify.MachineID)] = struct{}{}
					peer2 := &Peer{
						finder:   s,
						address:  peer.Address,
						identify: identify,
					}
					peers = append(peers, peer2)
					if notify != nil {
						notify(peer2)
					}
				}
			}
		},
	}); err != nil {
		return nil, err
	}

	return peers, nil
}

func (s *Discovery) Unicast(ctx context.Context, target *spec.Identify, name string, data []byte) ([]byte, error) {
	seq := xid.New().Bytes()
	req, err := rpc.Compress(data)
	if err != nil {
		return nil, err
	}
	req2 := &spec.Request{
		Target: &spec.Identify{MachineID: target.MachineID},
		Xid:    seq,
		Name:   name,
		Data:   req,
	}
	payload, err := proto.Marshal(req2)
	if err != nil {
		return nil, err
	}

	var resp []byte
	var err2 error

	once := sync.Once{}
	stopChan := make(chan struct{})

	reply := func(r []byte, e error) {
		resp = r
		err2 = e
		once.Do(func() {
			close(stopChan)
		})
	}

	go func() {
		select {
		case <-ctx.Done():
			reply(nil, errors.New("timeout"))
		}
	}()

	_, err = lan.Discover(lan.Settings{
		Port:      s.opts.Port,
		TimeLimit: -1,
		StopChan:  stopChan,
		Payload:   encodePacket(REQUEST, payload),
		Notify: func(peer lan.Discovered) {
			cmd, data, err := decodePacket(peer.Payload)
			if err != nil {
				s.logger.With(zap.Error(err)).Debug("invalid packet")
				return
			}

			if cmd == REPLY {
				resp := &spec.Response{}
				if err := proto.Unmarshal(data, resp); err != nil {
					reply(nil, err)
					return
				}

				if !bytes.Equal(seq, resp.GetXid()) {
					s.logger.With().Info("seq not match .. skipped")
					return
				}

				if resp.Code == 0 {
					reply(rpc.Decompress(resp.Data))
				} else {
					reply(nil, errors.New(resp.Msg))
				}
			}
		},
	})

	if err != nil {
		return nil, err
	}

	return resp, err2
}
