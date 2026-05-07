# Contracts

This directory holds the **shared event contracts** that define the boundary between services in LinkFlow v2.

## What's here

```
contracts/
├── CONVENTIONS.md       Naming, versioning, and evolution rules
├── MISTAKES.md          Design issues found before the first contract cleanup
└── events/              Event payload schemas (one file per event type)
    ├── envelope.schema.json
    ├── device.property.reported.v1.schema.json
    └── device.property.set.acknowledged.v1.schema.json
```

## Who reads what

| File type | Read by |
|---|---|
| `CONVENTIONS.md` | Humans — designers, reviewers, future contributors |
| `*.schema.json` | Programs — runtime validators, code generators, test tools |

The schemas are loaded at runtime by every service that produces or consumes events. They are not documentation; they are the actual contract enforced in code.

## Validation flow

Every producer and consumer MUST validate events in two stages:

1. Validate the full event object against `events/envelope.schema.json`.
2. Resolve the payload schema by `event_type + event_version`.
3. Validate `payload` against the resolved payload schema.

For example:

| `event_type` | `event_version` | Payload schema |
|---|---:|---|
| `device.property.reported` | `1` | `events/device.property.reported.v1.schema.json` |
| `device.property.set.acknowledged` | `1` | `events/device.property.set.acknowledged.v1.schema.json` |

Envelope validation alone is not enough. A valid envelope with the wrong payload is still an invalid event.

## Adding a new event

1. Read `CONVENTIONS.md` to confirm the naming rules.
2. Create `events/<domain>.<resource>.<action>.v1.schema.json`.
3. Validate the schema by writing example events that should pass and fail.
4. Update producers and consumers to load the new schema.

## Versioning rule

Schema version (`event_version` in the envelope) increments on every schema change by creating a new schema file version. The subject's major version (`lf.v1.*`) only changes on breaking changes that existing consumers cannot ignore. See `CONVENTIONS.md` for the full policy.

## Cross-language usage

Schemas are intentionally language-neutral. Each service loads them with its own validator:

- **Go**: `github.com/santhosh-tekuri/jsonschema/v6`
- **Python**: `jsonschema`
- **Rust**: `jsonschema` crate
- **TypeScript**: `ajv`

A single source of truth, enforced consistently across the stack.
