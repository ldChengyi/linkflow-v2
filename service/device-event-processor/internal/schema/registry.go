package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

type Registry struct {
	envelope *jsonschema.Schema
	payloads map[string]*jsonschema.Schema
}

type payloadSchemaSpec struct {
	EventType string
	Version   int
	File      string
}

// NewRegistry loads all event schemas from contractsFS.
//
// contractsFS should be rooted at the repository's contracts directory.
// For example:
//
//	os.DirFS("contracts")
//
// or, from service/device-event-processor:
//
//	os.DirFS("../../contracts")
func NewRegistry(contractsFs fs.FS) (*Registry, error) {
	complier := jsonschema.NewCompiler()

	// Draft 2020-12 disables format assertions by default in many cases.
	// The envelope uses uuid and date-time formats, so consumers should assert them]

	complier.AssertFormat()

	files := []string{
		event.EnvelopeSchemaFile,
		event.TelemetryReceivedV1SchemaFile,
	}

	for _, file := range files {
		doc, err := loadSchema(contractsFs, file)
		if err != nil {
			return nil, err
		}
		if err := complier.AddResource(file, doc); err != nil {
			return nil, fmt.Errorf("add schema resource %q: %w", file, err)
		}
	}

	envelopeSchema, err := complier.Compile(event.EnvelopeSchemaFile)
	if err != nil {
		return nil, fmt.Errorf("compile envelope schema: %w", err)
	}

	payloads := make(map[string]*jsonschema.Schema)
	for _, spec := range payloadSchemaSpecs() {
		schema, err := complier.Compile(spec.File)
		if err != nil {
			return nil, fmt.Errorf("compile payload schema %q: %w", spec.File, err)
		}
		payloads[event.PayloadSchemaKey(spec.EventType, spec.Version)] = schema
	}

	return &Registry{
		envelope: envelopeSchema,
		payloads: payloads,
	}, nil

}

func loadSchema(schemaFS fs.FS, file string) (any, error) {
	b, err := fs.ReadFile(schemaFS, file)
	if err != nil {
		return nil, fmt.Errorf("read schema %q: %w", file, err)
	}

	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("decode schema %q: %w", file, err)
	}
	return doc, nil
}

func payloadSchemaSpecs() []payloadSchemaSpec {
	return []payloadSchemaSpec{
		{
			EventType: event.TypeDeviceTelemetryReceived,
			Version:   event.VersionDeviceTelemetryReceived,
			File:      event.TelemetryReceivedV1SchemaFile,
		},
	}
}

func (r *Registry) ValidateEnvelope(raw []byte) error {
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("decode envelop json: %w", err)
	}

	if err := r.envelope.Validate(instance); err != nil {
		return fmt.Errorf("validate envelope: %w", err)
	}
	return nil
}

func (r *Registry) ValidatePayload(eventType string, version int, raw json.RawMessage) error {
	key := event.PayloadSchemaKey(eventType, version)

	schema, ok := r.payloads[key]
	if !ok {
		return fmt.Errorf("unknow payload schema for %s", key)
	}

	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("decode payload json: %w", err)
	}

	if err := schema.Validate(instance); err != nil {
		return fmt.Errorf("validate payload %s: %w", key, err)
	}

	return nil

}

// ValidateEvent validates the full event envelope first, then validates payload
// by event_type + event_version. It returns the decoded envelope after both
// schema checks pass.
func (r *Registry) ValidateEvent(raw []byte) (*event.Envelope, error) {
	if err := r.ValidateEnvelope(raw); err != nil {
		return nil, err
	}

	var env event.Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("decode envelope: %w", err)
	}

	if err := r.ValidatePayload(env.EventType, env.EventVersion, env.Payload); err != nil {
		return nil, err
	}

	return &env, nil
}
