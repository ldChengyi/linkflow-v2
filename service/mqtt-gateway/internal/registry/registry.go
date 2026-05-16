package registry

import (
	"fmt"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/handler"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/publisher"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/router"
)

type RouteSpec struct {
	Event      event.Spec
	MQTTFilter string
	QOS        byte
	Pattern    string
	Handler    router.Handler
}

type Subscription struct {
	Topic string
	QOS   byte
}

func Routes(ef event.EnvelopeFactory, pub publisher.Publisher) []RouteSpec {
	return []RouteSpec{
		{
			Event:      event.DevicePropertyReported,
			MQTTFilter: "lf/v1/+/+/property/up/post",
			QOS:        1,
			Pattern:    `^lf/v1/(?P<product_key>[^/]+)/(?P<device_slug>[^/]+)/property/up/post$`,
			Handler:    handler.DevicePropertyPost(event.DevicePropertyReported, ef, pub),
		},
		{
			Event:      event.DevicePropertySetAcknowledged,
			MQTTFilter: "lf/v1/+/+/property/up/set_reply",
			QOS:        1,
			Pattern:    `^lf/v1/(?P<product_key>[^/]+)/(?P<device_slug>[^/]+)/property/up/set_reply$`,
			Handler:    handler.DevicePropertySetReply(event.DevicePropertySetAcknowledged, ef, pub),
		},
	}
}

func RegisterAll(r *router.Router, ef event.EnvelopeFactory, pub publisher.Publisher, mws ...router.Middleware) ([]Subscription, error) {
	routes := Routes(ef, pub)
	subscriptions := make([]Subscription, 0, len(routes))
	for _, rt := range routes {
		if err := r.Handle(rt.Event.Key(), rt.Pattern, rt.Handler, mws...); err != nil {
			return nil, fmt.Errorf("register route for event %q: %w", rt.Event.Key(), err)
		}
		subscriptions = append(subscriptions, Subscription{
			Topic: rt.MQTTFilter,
			QOS:   rt.QOS,
		})
	}
	return subscriptions, nil
}
