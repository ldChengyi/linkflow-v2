# Event Conventions

This document defines naming, versioning, and evolution rules for all events in LinkFlow v2.

Every event schema in this directory MUST follow these conventions. Conventions are stable; changing them requires updating this document and reviewing all affected schemas.

## Subject Naming

All event subjects follow this pattern:

```
lf.<version>.<domain>.<resource>.<action>
```

Examples:

```
lf.v1.device.telemetry.received
lf.v1.device.status.changed
lf.v1.device.command.requested
lf.v1.alert.lifecycle.created
```

Rules:

- `lf` — fixed project prefix (LinkFlow). Reserved.
- `<version>` — major version of the event family, e.g. `v1`, `v2`. See [Versioning](#versioning).
- `<domain>` — top-level business domain: `device`, `alert`, `rule`, `ota`, `user`.
- `<resource>` — sub-entity inside the domain: `telemetry`, `status`, `command`.
- `<action>` — what happened, in past tense passive voice: `received`, `changed`, `created`, `requested`.

For domains without a natural sub-resource, use a stable resource name instead of shortening the subject. For example, use `lf.v1.alert.lifecycle.created`, not `lf.v1.alert.created`.

Subjects MUST NOT contain:

- tenant identifiers (`tenant_id` belongs in the payload)
- device identifiers (`device_id` belongs in the payload)
- environment names (`prod`, `staging` — handled by deployment, not subject)
- timestamps

## Action Naming

Actions describe facts that have already happened. Never use forward-looking or imperative verbs.

Good:

- `received` — telemetry has been received from a device
- `changed` — a status transition has occurred
- `created` — a new entity exists
- `requested` — a command has been requested by some actor
- `dispatched` — a command has been sent to the device
- `acknowledged` — the device confirmed the command
- `failed` — an operation has failed

Bad:

- `receive` (imperative) — events are not commands
- `will_change` (future tense) — events describe the past
- `process` (vague action) — name the actual outcome

Each state transition should produce its own event. Do not collapse multiple transitions into one event.

## Field Naming

- All field names use `snake_case`.
- Time fields end in `_at` (e.g. `occurred_at`, `created_at`).
- ID fields end in `_id` (e.g. `event_id`, `device_id`, `tenant_id`).
- Boolean fields are positive statements, not negations (`is_active`, not `is_not_inactive`).

## Time Format

All time fields use **RFC 3339 / ISO 8601** with timezone:

```
2026-05-02T10:30:00Z
2026-05-02T10:30:00.123Z
2026-05-02T18:30:00+08:00
```

UTC (`Z` suffix) is preferred. Local timezones are allowed but discouraged.

## Required vs Optional Fields

A field is **required** if a consumer cannot reasonably process the event without it. Everything else is optional.

Required fields in every event envelope:

- `event_id`
- `event_type`
- `event_version`
- `occurred_at`
- `producer`
- `tenant_id`
- `payload`

Optional but recommended:

- `trace_id` — required if observability is enabled
- `correlation_id` — groups events caused by the same external request or workflow
- `causation_id` — points to the event that directly caused this event

The payload schema decides which payload fields are required.

## Validation

Every event MUST be validated in two stages:

1. Validate the complete event object against `events/envelope.schema.json`.
2. Resolve the payload schema by `event_type + event_version`.
3. Validate the `payload` object against the resolved payload schema.

For example:

| `event_type` | `event_version` | Payload schema |
|---|---:|---|
| `device.telemetry.received` | `1` | `events/device.telemetry.received.v1.schema.json` |

Envelope validation alone is not sufficient. It proves the common metadata shape, but it does not prove that the payload matches the event type.

## Versioning

The version in the subject (`v1`, `v2`) is the **major version** of the event family. It changes only on **breaking changes**.

A change is breaking if it can cause an existing consumer to fail. Examples of breaking changes:

- removing a required field
- renaming a field
- changing a field type (e.g. string to int)
- changing the meaning of a field
- changing enum values that consumers may match against

A change is non-breaking if existing consumers can ignore it safely. Examples of non-breaking changes:

- adding a new optional field
- adding a new event type
- adding a new value to an open-ended enum

Every schema change creates a new schema file version. Do not rewrite a released schema in place in a way that changes its accepted contract.

Non-breaking changes increment the schema's `event_version` integer and create a new payload schema file, but the subject stays at `v1`.

Breaking changes require a new subject (`lf.v2.device.telemetry.received`) and the old subject continues to be supported until all consumers migrate.

## Schema Files

Each event has a JSON Schema file under `events/`:

```
events/
  envelope.schema.json
  device.telemetry.received.v1.schema.json
  device.status.changed.v1.schema.json
```

File naming: `<domain>.<resource>.<action>.v<N>.schema.json`

Each schema MUST:

- declare `$schema` at the top (use `https://json-schema.org/draft/2020-12/schema`)
- declare a `$id` matching its canonical contract identity
- include a `description` field explaining when this event is produced
- list `required` fields explicitly

Canonical payload schema `$id` format:

```
https://linkflow.dev/contracts/events/<subject>
```

Example:

```
https://linkflow.dev/contracts/events/lf.v1.device.telemetry.received
```

## Idempotency

Every event has a globally unique `event_id` (UUID v7 preferred for time-ordered keys, UUID v4 acceptable). Consumers MUST treat duplicate `event_id` values as no-ops.

Producers should ensure that retrying the same business operation produces events with the same `event_id`. This makes the entire pipeline safely retryable.

## Tenant Isolation

Every event carries `tenant_id` in the envelope. Consumers MUST filter by `tenant_id` when storing or querying data. The current implementation may use `"default"` as a placeholder until multi-tenancy is fully enabled, but the field is mandatory from day one.

## Trace Propagation

When `trace_id` is present, it follows the W3C Trace Context format. Producers should propagate the active trace context into the event; consumers should resume the trace when handling the event.

## Examples

A minimal valid envelope:

```json
{
  "event_id": "018f0a6e-7b3e-7a7b-9c2e-9b9a0c9b9f10",
  "event_type": "device.telemetry.received",
  "event_version": 1,
  "occurred_at": "2026-05-02T10:30:00Z",
  "producer": "ingest-gateway",
  "tenant_id": "default",
  "payload": {}
}
```

With observability:

```json
{
  "event_id": "018f0a6e-7b3e-7a7b-9c2e-9b9a0c9b9f10",
  "event_type": "device.telemetry.received",
  "event_version": 1,
  "occurred_at": "2026-05-02T10:30:00Z",
  "producer": "ingest-gateway",
  "tenant_id": "default",
  "trace_id": "0af7651916cd43dd8448eb211c80319c",
  "payload": {}
}
```
