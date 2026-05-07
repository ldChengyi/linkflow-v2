package validation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

type Validator struct {
	envelope *jsonschema.Schema
	payloads map[string]*jsonschema.Schema
}

// NewValidator loads all event schemas from contractsFS.
//
// contractsFS should be rooted at the repository's contracts directory.
// For example:
//
//	os.DirFS("contracts")
//
// or, from service/device-event-processor:
//
//	os.DirFS("../../contracts")
func NewValidator(contractsFs fs.FS) (*Validator, error) {
	compiler := jsonschema.NewCompiler()

	// Draft 2020-12 disables format assertions by default in many cases.
	// The envelope uses uuid and date-time formats, so consumers should assert them.

	compiler.AssertFormat()

	files := []string{event.EnvelopeSchemaFile}
	for _, spec := range event.Specs() {
		files = append(files, spec.SchemaFile)
	}

	for _, file := range files {
		doc, err := loadSchema(contractsFs, file)
		if err != nil {
			return nil, err
		}
		if err := compiler.AddResource(file, doc); err != nil {
			return nil, fmt.Errorf("add schema resource %q: %w", file, err)
		}
	}

	envelopeSchema, err := compiler.Compile(event.EnvelopeSchemaFile)
	if err != nil {
		return nil, fmt.Errorf("compile envelope schema: %w", err)
	}

	payloads := make(map[string]*jsonschema.Schema)
	for _, spec := range event.Specs() {
		schema, err := compiler.Compile(spec.SchemaFile)
		if err != nil {
			return nil, fmt.Errorf("compile payload schema %q: %w", spec.SchemaFile, err)
		}
		payloads[spec.Key()] = schema
	}

	return &Validator{
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

func (v *Validator) ValidateEnvelope(raw []byte) error {
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("decode envelope json: %w", err)
	}

	if err := v.envelope.Validate(instance); err != nil {
		return fmt.Errorf("validate envelope: %w", err)
	}
	return nil
}

func (v *Validator) ValidatePayload(eventType string, version int, raw json.RawMessage) error {
	key := event.PayloadSchemaKey(eventType, version)

	schema, ok := v.payloads[key]
	if !ok {
		return fmt.Errorf("unknown payload schema for %s", key)
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
func (v *Validator) ValidateEvent(raw []byte) (*event.Envelope, error) {
	if err := v.ValidateEnvelope(raw); err != nil {
		return nil, err
	}

	var env event.Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("decode envelope: %w", err)
	}

	if err := v.ValidatePayload(env.EventType, env.EventVersion, env.Payload); err != nil {
		return nil, err
	}

	return &env, nil
}
