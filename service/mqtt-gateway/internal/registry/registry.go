package registry

import (
	"fmt"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/handler"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/publisher"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/router"
)

type RouteSpec struct {
	Name    string
	Pattern string
	Handler router.Handler
}

func RegisterAll(r *router.Router, ef event.EnvelopeFactory, pub publisher.Publisher) error {
	routes := []RouteSpec{
		{
			Name:    "device.property.post",
			Pattern: `^lf/v1/(?P<product_key>[^/]+)/(?P<device_id>[^/]+)/property/up/post$`,
			Handler: handler.Property(ef, pub),
		},
	}

	for _, rt := range routes {
		if err := r.Handle(rt.Name, rt.Pattern, rt.Handler); err != nil {
			return fmt.Errorf("register route %q: %w", rt.Name, err)
		}
	}
	return nil
}
