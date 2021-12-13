package dnet

import (
	"bytes"
	"io"
	"strings"

	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"

	"github.com/cima-go/dnet/lan"
	"github.com/cima-go/dnet/rpc"
	"github.com/cima-go/dnet/spec"
)

func Advertise(opts Options, idx *spec.Identify, rpc *rpc.Registry, logger *zap.Logger) io.Closer {
	logger = logger.With(zap.String("mod", "advertise"))

	stopChan := make(chan struct{})
	waitChan := make(chan struct{})

	go func() {
		peers, err := lan.Discover(lan.Settings{
			MulticastAddress: opts.Addr,
			Port:             opts.Port,
			TimeLimit:        -1,
			StopChan:         stopChan,
			DisableBroadcast: true,
			Notify: func(peer lan.Discovered) {
				cmd, data, err := decodePacket(peer.Payload)
				if err != nil {
					logger.With(zap.Error(err)).Debug("invalid packet")
					return
				}

				switch cmd {
				case KNOCKING:
					logger.With(zap.String("src", peer.Address)).Debug("knocking")
					knockResponse(peer, idx)
				case REQUEST:
					req := &spec.Request{}
					if err := proto.Unmarshal(data, req); err != nil {
						logger.With(zap.Error(err)).Info("message decode failed")
					}

					if !bytes.Equal(idx.MachineID, req.Target.MachineID) {
						logger.With(zap.String("dst", req.Target.Hostname)).Info("target not match .. skipped")
						return
					}

					if err := rpc.Dispatch(req.GetXid(), req.GetName(), req.GetData(), func(data []byte, err error) {
						resp := &spec.Response{Xid: req.Xid}
						if err == nil {
							resp.Code = 0
							resp.Data = data
						} else {
							resp.Code = 1
							resp.Msg = strings.ToValidUTF8(err.Error(), "*")
						}
						if payload, err := proto.Marshal(resp); err == nil {
							peer.Reply(encodePacket(REPLY, payload))
						} else {
							logger.With(zap.Error(err)).Info("resp marshal failed")
						}
					}); err != nil {
						logger.With(zap.Error(err)).Info("rpc dispatch failed")
					}
				}
			},
		})

		if err != nil {
			logger.With(zap.Error(err)).Info("advertise failed")
		} else {
			logger.With(zap.Int("foundPeers", len(peers))).Debug("advertise stopped")
		}

		close(waitChan)
	}()

	return &Advertiser{stopChan, waitChan}
}

type Advertiser struct {
	stopChan chan struct{}
	waitChan chan struct{}
}

func (a *Advertiser) Close() error {
	close(a.stopChan)
	<-a.waitChan
	return nil
}
