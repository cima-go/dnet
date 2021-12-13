package rpc

import (
	"sync"
	"time"

	"github.com/bluele/gcache"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func Init(logger *zap.Logger) *Registry {
	return &Registry{
		dedup:  gcache.New(1024).Simple().Expiration(time.Minute).Build(),
		logger: logger,
	}
}

type Registry struct {
	dedup  gcache.Cache
	routes sync.Map
	logger *zap.Logger
}

type Handler func(req []byte) (resp []byte, err error)

func (s *Registry) Register(name string, handler Handler) error {
	if _, exists := s.routes.LoadOrStore(name, handler); exists {
		return errors.New("name exists")
	}
	return nil
}

func (s *Registry) Dispatch(seq []byte, name string, data []byte, reply func([]byte, error)) error {
	handler, exists := s.routes.Load(name)
	if !exists {
		return errors.New("service not found")
	}

	if s.dedup.Has(string(seq)) {
		s.logger.With(zap.String("service", name)).Debug("duplicated request")
		return nil
	}

	_ = s.dedup.Set(string(seq), 1)

	req, err := Decompress(data)
	if err != nil {
		reply(nil, err)
		return nil
	}

	resp, err := handler.(Handler)(req)
	if err != nil {
		reply(nil, err)
		return nil
	}

	reply(Compress(resp))
	return nil
}
