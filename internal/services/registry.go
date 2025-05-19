package services

import (
	"log/slog"

	"github.com/puzpuzpuz/xsync/v3"
	"golang.org/x/sync/errgroup"
)

type Registry struct {
	services *xsync.MapOf[string, Service]
}

type ServiceName string

const (
	OLSRServiceName    ServiceName = "olsr"
	BabelServiceName   ServiceName = "babel"
	DNSMasqServiceName ServiceName = "dnsmasq"
)

func NewServiceRegistry() *Registry {
	return &Registry{
		services: xsync.NewMapOf[string, Service](),
	}
}

func (r *Registry) Register(name ServiceName, service Service) {
	r.services.Store(string(name), service)
}

func (r *Registry) Get(name ServiceName) (Service, bool) {
	return r.services.Load(string(name))
}

func (r *Registry) StartAll() {
	r.services.Range(func(name string, service Service) bool {
		if !service.IsEnabled() {
			slog.Debug("service is disabled", "service", name)
			return true
		}
		go func() error {
			for {
				err := service.Start()
				if err != nil {
					slog.Warn("service failed to start", "service", name, "error", err)
				}
			}
		}()
		return true
	})
}

func (r *Registry) StopAll() error {
	errGrp := errgroup.Group{}
	r.services.Range(func(name string, service Service) bool {
		errGrp.Go(func() error {
			return service.Stop()
		})
		return true
	})
	return errGrp.Wait()
}
