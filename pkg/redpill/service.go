package redpill

import (
	"sync"
)

var init_schema sync.Once

type serviceImpl struct {
	options Options
	lock    sync.Mutex
	stop    chan bool
}

type Service interface {
	Stop()
	Close() error
}

func NewService(options Options) (Service, error) {
	impl := &serviceImpl{
		stop: make(chan bool),
	}

	return impl, nil
}

func (this *serviceImpl) Stop() {
	this.stop <- true
}

func (this *serviceImpl) Close() error {
	this.stop <- true
	return nil
}
