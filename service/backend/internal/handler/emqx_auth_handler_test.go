package handler

import (
	"log/slog"
	"testing"
)

func TestMQTTAuthInputFromEMQXSupportsTenantUsernameAndDeviceClientID(t *testing.T) {
	in, ok := mqttAuthInputFromEMQX(emqxAuthRequest{
		Username: " default ",
		Password: " device-secret ",
		ClientID: " device-01 ",
	})
	if !ok {
		t.Fatal("mqttAuthInputFromEMQX() ok = false")
	}
	if in.TenantSlug != "default" || in.ProductKey != "" || in.DeviceSlug != "device-01" || in.Password != "device-secret" {
		t.Fatalf("input = %+v, want tenant default device device-01", in)
	}
}

func TestMQTTAuthInputFromEMQXSupportsFullUsername(t *testing.T) {
	in, ok := mqttAuthInputFromEMQX(emqxAuthRequest{
		Username: "default:product-01:device-01",
		Password: "device-secret",
		ClientID: "mqtt-client-01",
	})
	if !ok {
		t.Fatal("mqttAuthInputFromEMQX() ok = false")
	}
	if in.TenantSlug != "default" || in.ProductKey != "product-01" || in.DeviceSlug != "device-01" {
		t.Fatalf("input = %+v, want full identity from username", in)
	}
}

func TestMQTTAuthInputFromEMQXRejectsInvalidIdentity(t *testing.T) {
	if _, ok := mqttAuthInputFromEMQX(emqxAuthRequest{
		Username: "default:too:many:parts",
		Password: "device-secret",
		ClientID: "device-01",
	}); ok {
		t.Fatal("mqttAuthInputFromEMQX() ok = true, want false")
	}
}

func TestEMQXAuthHandlerAuthenticatesServiceClient(t *testing.T) {
	h := &EMQXAuthHandler{
		serviceUsername: "linkflow-mqtt-gateway",
		servicePassword: "linkflow-mqtt-gateway-secret",
		log:             slog.Default(),
	}

	if !h.authenticateServiceClient(emqxAuthRequest{
		Username: "linkflow-mqtt-gateway",
		Password: "linkflow-mqtt-gateway-secret",
		ClientID: "linkflow-mqtt-gateway-fullstack-test",
	}) {
		t.Fatal("authenticateServiceClient() = false, want true")
	}
	if h.authenticateServiceClient(emqxAuthRequest{
		Username: "linkflow-mqtt-gateway",
		Password: "wrong",
		ClientID: "linkflow-mqtt-gateway-fullstack-test",
	}) {
		t.Fatal("authenticateServiceClient() = true, want false")
	}
}
