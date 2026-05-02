# LinkFlow v2 — Event-Driven IoT Cloud Platform

> An IoT cloud platform built on event-driven architecture, with polyglot services and enterprise-grade middleware.

[简体中文](./README_zh-CN.md) | English

## Background

LinkFlow v1 was a Go monolith with a process-driven architecture. As features grew, it accumulated significant coupling: modules called each other directly, services shared the same database tables, and protocol handling was tangled with business logic. LinkFlow v2 is a ground-up redesign around an event bus and clear domain boundaries.

## Core Design

- **Event-Driven Architecture (EDA)** — services communicate asynchronously through a Kafka event bus, with no direct dependencies between them.
- **Polyglot services** — each language is used where it shines: Go for business logic, Python for AI, Rust for performance-sensitive paths.
- **Domain-Driven Design (DDD)** — services are split by bounded contexts, each with its own database.
- **Observability-first** — distributed tracing flows across all languages and services.

## Tech Stack

### Infrastructure
- **MQTT Broker**: EMQX 5.x
- **Event Bus**: Redpanda (Kafka API compatible)
- **Relational DB**: PostgreSQL 16
- **Time-series DB**: TimescaleDB 2.x
- **Cache**: Redis 7.x

### Application Services
- **Business services**: Go + Kratos
- **AI service**: Python + FastAPI
- **Ingestion layer**: Rust + Tokio
- **Frontend**: React + TypeScript + Ant Design Pro

## Branching Model

- `main` — release branch. Every commit corresponds to a tagged, runnable version.
- `dev` — active development branch. Feature work happens here and is merged into `main` via pull request when a milestone is reached.

## License

This repository is part of an undergraduate capstone project.
