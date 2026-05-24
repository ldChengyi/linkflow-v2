package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/service"
)

type PostgresDeviceStore struct {
	actor actorRLSStore
}

func NewPostgresDeviceStore(pool *pgxpool.Pool) (*PostgresDeviceStore, error) {
	if pool == nil {
		return nil, fmt.Errorf("postgres pool is nil")
	}
	return &PostgresDeviceStore{actor: newActorRLSStore(pool, "device")}, nil
}

func (s *PostgresDeviceStore) FindDeviceProductAuthType(ctx context.Context, in service.DeviceProductAuthInput) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	const query = `SELECT auth_type FROM products WHERE id = $1 AND tenant_id = $2`

	var authType string
	err := s.withDeviceUser(ctx, in.UserID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query, in.ProductID, in.TenantID).Scan(&authType)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", service.ErrInvalidDeviceInput
		}
		return "", fmt.Errorf("find device product auth type: %w", err)
	}
	return authType, nil
}

func (s *PostgresDeviceStore) FindMQTTDeviceCredential(ctx context.Context, in service.MQTTAuthInput) (service.MQTTDeviceCredential, error) {
	if err := ctx.Err(); err != nil {
		return service.MQTTDeviceCredential{}, err
	}

	const query = `
SELECT
    t.id::text,
    p.id::text,
    d.id::text,
    t.tenant_slug,
    p.product_key,
    d.device_slug,
    p.auth_type,
    t.status,
    p.status,
    d.status,
    COALESCE(dc.status, ''),
    COALESCE(dc.secret_hash, '')
FROM tenants t
JOIN products p ON p.tenant_id = t.id
JOIN devices d ON d.tenant_id = t.id AND d.product_id = p.id
LEFT JOIN device_credentials dc ON dc.tenant_id = t.id AND dc.device_id = d.id AND dc.status = 'active'
WHERE lower(t.tenant_slug) = $1
  AND ($2 = '' OR lower(p.product_key) = $2)
  AND lower(d.device_slug) = $3`

	var cred service.MQTTDeviceCredential
	err := s.actor.withInternalService(ctx, "backend", func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query, in.TenantSlug, in.ProductKey, in.DeviceSlug).Scan(
			&cred.TenantID,
			&cred.ProductID,
			&cred.DeviceID,
			&cred.TenantSlug,
			&cred.ProductKey,
			&cred.DeviceSlug,
			&cred.ProductAuthType,
			&cred.TenantStatus,
			&cred.ProductStatus,
			&cred.DeviceStatus,
			&cred.CredentialStatus,
			&cred.SecretHash,
		)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.MQTTDeviceCredential{}, service.ErrDeviceNotFound
		}
		return service.MQTTDeviceCredential{}, fmt.Errorf("find mqtt device credential: %w", err)
	}
	return cred, nil
}

func (s *PostgresDeviceStore) CreateDevice(ctx context.Context, in service.DeviceCreateInput) (service.Device, error) {
	if err := ctx.Err(); err != nil {
		return service.Device{}, err
	}

	const query = `
INSERT INTO devices (
    tenant_id,
    product_id,
    device_slug,
    device_name,
    description,
    gateway_device_id
) VALUES (
    $1, $2, $3, $4, $5, $6
)
	RETURNING id::text, tenant_id::text, product_id::text, device_slug, device_name, description, status, connection_status, COALESCE(gateway_device_id::text, ''), COALESCE(firmware_version, ''), COALESCE(ip_address, ''), last_seen_at, created_at, updated_at`

	var device service.Device
	err := s.withDeviceUser(ctx, in.UserID, func(tx pgx.Tx) error {
		ok, err := productBelongsToTenant(ctx, tx, in.ProductID, in.TenantID)
		if err != nil {
			return err
		}
		if !ok {
			return service.ErrInvalidDeviceInput
		}
		if in.GatewayDeviceID != "" {
			ok, err := deviceBelongsToTenant(ctx, tx, in.GatewayDeviceID, in.TenantID)
			if err != nil {
				return err
			}
			if !ok {
				return service.ErrInvalidDeviceInput
			}
		}

		if err := scanDevice(tx.QueryRow(
			ctx,
			query,
			in.TenantID,
			in.ProductID,
			in.DeviceSlug,
			in.DeviceName,
			in.Description,
			nullIfEmpty(in.GatewayDeviceID),
		), &device); err != nil {
			return err
		}

		if in.DeviceSecretHash != "" {
			if err := insertActiveDeviceCredential(ctx, tx, device.TenantID, device.ID, in.DeviceSecretHash); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if isUniqueViolation(err) {
			return service.Device{}, service.ErrDeviceAlreadyExists
		}
		return service.Device{}, fmt.Errorf("insert device: %w", err)
	}
	return device, nil
}

func (s *PostgresDeviceStore) ListDevices(ctx context.Context, in service.DeviceListInput) (service.PageResult[service.Device], error) {
	if err := ctx.Err(); err != nil {
		return service.PageResult[service.Device]{}, err
	}

	const query = `
SELECT id::text, tenant_id::text, product_id::text, device_slug, device_name, description, status, connection_status, COALESCE(gateway_device_id::text, ''), COALESCE(firmware_version, ''), COALESCE(ip_address, ''), last_seen_at, created_at, updated_at
FROM devices
WHERE tenant_id = $1
  AND ($2 = '' OR product_id::text = $2)
ORDER BY created_at DESC, id DESC
LIMIT $3 OFFSET $4`

	const countQuery = `
SELECT count(*)
FROM devices
WHERE tenant_id = $1
  AND ($2 = '' OR product_id::text = $2)`

	var devices []service.Device
	var total int
	err := s.withDeviceUser(ctx, in.UserID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, countQuery, in.TenantID, in.ProductID).Scan(&total); err != nil {
			return err
		}

		rows, err := tx.Query(ctx, query, in.TenantID, in.ProductID, in.PageInput.Limit(), in.PageInput.Offset())
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var device service.Device
			if err := scanDevice(rows, &device); err != nil {
				return err
			}
			devices = append(devices, device)
		}
		return rows.Err()
	})
	if err != nil {
		return service.PageResult[service.Device]{}, fmt.Errorf("list devices: %w", err)
	}
	return service.NewPageResult(devices, total, in.PageInput), nil
}

func (s *PostgresDeviceStore) FindDeviceByID(ctx context.Context, in service.DeviceGetInput) (service.Device, error) {
	if err := ctx.Err(); err != nil {
		return service.Device{}, err
	}

	const query = `
SELECT id::text, tenant_id::text, product_id::text, device_slug, device_name, description, status, connection_status, COALESCE(gateway_device_id::text, ''), COALESCE(firmware_version, ''), COALESCE(ip_address, ''), last_seen_at, created_at, updated_at
FROM devices
WHERE id = $1`

	var device service.Device
	err := s.withDeviceUser(ctx, in.UserID, func(tx pgx.Tx) error {
		return scanDevice(tx.QueryRow(ctx, query, in.DeviceID), &device)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.Device{}, service.ErrDeviceNotFound
		}
		return service.Device{}, fmt.Errorf("find device by id: %w", err)
	}
	return device, nil
}

func (s *PostgresDeviceStore) FindDeviceLatestProperties(ctx context.Context, in service.DeviceLatestPropertiesInput) (service.DeviceLatestProperties, error) {
	if err := ctx.Err(); err != nil {
		return service.DeviceLatestProperties{}, err
	}

	const query = `
SELECT
    d.id::text,
    d.tenant_id::text,
    d.product_id::text,
    p.product_key,
    d.device_slug,
    latest.properties,
    latest.occurred_at,
    latest.received_at
FROM devices d
JOIN products p ON p.id = d.product_id AND p.tenant_id = d.tenant_id
LEFT JOIN LATERAL (
    SELECT properties, occurred_at, received_at
    FROM device_property_report_events
    WHERE tenant_id = d.tenant_id::text
      AND product_key = p.product_key
      AND device_slug = d.device_slug
    ORDER BY occurred_at DESC
    LIMIT 1
) latest ON true
WHERE d.id = $1`

	var latest service.DeviceLatestProperties
	var properties []byte
	var occurredAt sql.NullTime
	var receivedAt sql.NullTime
	err := s.withDeviceUser(ctx, in.UserID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query, in.DeviceID).Scan(
			&latest.ID,
			&latest.TenantID,
			&latest.ProductID,
			&latest.ProductKey,
			&latest.DeviceSlug,
			&properties,
			&occurredAt,
			&receivedAt,
		)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.DeviceLatestProperties{}, service.ErrDeviceNotFound
		}
		return service.DeviceLatestProperties{}, fmt.Errorf("find device latest properties: %w", err)
	}

	latest.Properties = map[string]any{}
	if len(properties) > 0 {
		if err := json.Unmarshal(properties, &latest.Properties); err != nil {
			return service.DeviceLatestProperties{}, fmt.Errorf("unmarshal device latest properties: %w", err)
		}
		latest.Reported = true
	}
	if occurredAt.Valid {
		t := occurredAt.Time
		latest.OccurredAt = &t
	}
	if receivedAt.Valid {
		t := receivedAt.Time
		latest.ReceivedAt = &t
	}
	return latest, nil
}

func (s *PostgresDeviceStore) FindDevicePropertyTrend(ctx context.Context, in service.DevicePropertyTrendInput) (service.DevicePropertyTrend, error) {
	if err := ctx.Err(); err != nil {
		return service.DevicePropertyTrend{}, err
	}

	const scopeQuery = `
SELECT
    d.id::text,
    d.tenant_id::text,
    d.product_id::text,
    p.product_key,
    d.device_slug
FROM devices d
JOIN products p ON p.id = d.product_id AND p.tenant_id = d.tenant_id
WHERE d.id = $1`

	const trendQuery = `
WITH samples AS (
    SELECT
        prop.key AS property,
        time_bucket(make_interval(secs => $7::int), e.occurred_at) AS bucket_at,
        (prop.value #>> '{}')::double precision AS value,
        e.occurred_at
    FROM device_property_report_events e
    CROSS JOIN LATERAL jsonb_each(e.properties) AS prop(key, value)
    WHERE e.tenant_id = $1
      AND e.product_key = $2
      AND e.device_slug = $3
      AND prop.key = ANY($4::text[])
      AND jsonb_typeof(prop.value) = 'number'
      AND e.occurred_at >= $5
      AND e.occurred_at < $6
)
SELECT
    property,
    bucket_at,
    avg(value),
    min(value),
    max(value),
    last(value, occurred_at),
    count(*)
FROM samples
GROUP BY property, bucket_at
ORDER BY property ASC, bucket_at ASC`

	trend := service.DevicePropertyTrend{
		Properties:    append([]string(nil), in.Properties...),
		From:          in.From,
		To:            in.To,
		BucketSeconds: in.BucketSeconds,
		Aggregate:     in.Aggregate,
		Series:        make([]service.DevicePropertyTrendSeries, 0, len(in.Properties)),
	}
	seriesByProperty := make(map[string]int, len(in.Properties))
	for _, property := range in.Properties {
		seriesByProperty[property] = len(trend.Series)
		trend.Series = append(trend.Series, service.DevicePropertyTrendSeries{
			Property: property,
			Points:   []service.DevicePropertyTrendPoint{},
		})
	}

	err := s.withDeviceUser(ctx, in.UserID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, scopeQuery, in.DeviceID).Scan(
			&trend.DeviceID,
			&trend.TenantID,
			&trend.ProductID,
			&trend.ProductKey,
			&trend.DeviceSlug,
		); err != nil {
			return err
		}

		rows, err := tx.Query(
			ctx,
			trendQuery,
			trend.TenantID,
			trend.ProductKey,
			trend.DeviceSlug,
			in.Properties,
			in.From,
			in.To,
			in.BucketSeconds,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var (
				property  string
				bucketAt  sql.NullTime
				avgValue  float64
				minValue  float64
				maxValue  float64
				lastValue float64
				count     int
			)
			if err := rows.Scan(
				&property,
				&bucketAt,
				&avgValue,
				&minValue,
				&maxValue,
				&lastValue,
				&count,
			); err != nil {
				return err
			}
			if !bucketAt.Valid {
				continue
			}
			seriesIndex, ok := seriesByProperty[property]
			if !ok {
				continue
			}
			trend.Series[seriesIndex].Points = append(trend.Series[seriesIndex].Points, service.DevicePropertyTrendPoint{
				BucketAt: bucketAt.Time,
				Value:    trendAggregateValue(in.Aggregate, avgValue, minValue, maxValue, lastValue),
				Min:      minValue,
				Max:      maxValue,
				Count:    count,
			})
		}
		return rows.Err()
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.DevicePropertyTrend{}, service.ErrDeviceNotFound
		}
		return service.DevicePropertyTrend{}, fmt.Errorf("find device property trend: %w", err)
	}

	return trend, nil
}

func trendAggregateValue(aggregate string, avgValue float64, minValue float64, maxValue float64, lastValue float64) float64 {
	switch aggregate {
	case "min":
		return minValue
	case "max":
		return maxValue
	case "last":
		return lastValue
	default:
		return avgValue
	}
}

func (s *PostgresDeviceStore) ListDeviceEventHistory(ctx context.Context, in service.DeviceEventHistoryInput) (service.PageResult[service.DeviceEventEntry], error) {
	if err := ctx.Err(); err != nil {
		return service.PageResult[service.DeviceEventEntry]{}, err
	}

	const scopeQuery = `
SELECT
    d.tenant_id::text,
    d.product_id::text,
    p.product_key,
    d.device_slug
FROM devices d
JOIN products p ON p.id = d.product_id AND p.tenant_id = d.tenant_id
WHERE d.id = $1`

	const countQuery = `
SELECT count(*)
FROM device_event_report_events
WHERE tenant_id = $1
  AND product_key = $2
  AND device_slug = $3
  AND ($4 = '' OR event_name = $4)`

	const query = `
SELECT
    event_id::text,
    tenant_id,
    product_key,
    device_slug,
    event_name,
    params,
    occurred_at,
    received_at
FROM device_event_report_events
WHERE tenant_id = $1
  AND product_key = $2
  AND device_slug = $3
  AND ($4 = '' OR event_name = $4)
ORDER BY occurred_at DESC, event_id DESC
LIMIT $5 OFFSET $6`

	var (
		tenantID   string
		productID  string
		productKey string
		deviceSlug string
		events     []service.DeviceEventEntry
		total      int
	)
	err := s.withDeviceUser(ctx, in.UserID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, scopeQuery, in.DeviceID).Scan(
			&tenantID,
			&productID,
			&productKey,
			&deviceSlug,
		); err != nil {
			return err
		}

		if err := tx.QueryRow(ctx, countQuery, tenantID, productKey, deviceSlug, in.EventName).Scan(&total); err != nil {
			return err
		}

		rows, err := tx.Query(
			ctx,
			query,
			tenantID,
			productKey,
			deviceSlug,
			in.EventName,
			in.PageInput.Limit(),
			in.PageInput.Offset(),
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var entry service.DeviceEventEntry
			var params []byte
			if err := rows.Scan(
				&entry.EventID,
				&entry.TenantID,
				&entry.ProductKey,
				&entry.DeviceSlug,
				&entry.EventName,
				&params,
				&entry.OccurredAt,
				&entry.ReceivedAt,
			); err != nil {
				return err
			}
			entry.ProductID = productID
			entry.Params = map[string]any{}
			if len(params) > 0 {
				if err := json.Unmarshal(params, &entry.Params); err != nil {
					return fmt.Errorf("unmarshal device event params: %w", err)
				}
			}
			events = append(events, entry)
		}
		return rows.Err()
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.PageResult[service.DeviceEventEntry]{}, service.ErrDeviceNotFound
		}
		return service.PageResult[service.DeviceEventEntry]{}, fmt.Errorf("list device event history: %w", err)
	}

	return service.NewPageResult(events, total, in.PageInput), nil
}

func (s *PostgresDeviceStore) ListDeviceServiceCallHistory(ctx context.Context, in service.DeviceServiceCallHistoryInput) (service.PageResult[service.DeviceServiceCallHistoryEntry], error) {
	if err := ctx.Err(); err != nil {
		return service.PageResult[service.DeviceServiceCallHistoryEntry]{}, err
	}

	const scopeQuery = `
SELECT
    d.tenant_id::text,
    d.product_id::text,
    p.product_key,
    d.device_slug
FROM devices d
JOIN products p ON p.id = d.product_id AND p.tenant_id = d.tenant_id
WHERE d.id = $1`

	const countQuery = `
SELECT count(*)
FROM device_service_call_events
WHERE tenant_id = $1
  AND product_key = $2
  AND device_slug = $3
  AND ($4 = '' OR service_name = $4)`

	const query = `
SELECT
    c.command_id::text,
    c.tenant_id,
    c.product_key,
    c.device_slug,
    c.service_name,
    c.topic,
    c.input,
    c.occurred_at,
    c.occurred_at + make_interval(secs => $5::int),
    CASE
      WHEN a.event_id IS NULL AND now() <= c.occurred_at + make_interval(secs => $5::int) THEN 'pending'
      WHEN a.event_id IS NULL THEN 'failed'
      WHEN a.success THEN 'success'
      ELSE 'failed'
    END,
    COALESCE(a.event_id::text, ''),
    a.success,
    COALESCE(a.code, ''),
    COALESCE(a.message, ''),
    COALESCE(a.output, '{}'::jsonb),
    a.occurred_at,
    a.received_at
FROM device_service_call_events c
LEFT JOIN LATERAL (
    SELECT event_id, success, code, message, output, occurred_at, received_at
    FROM device_service_call_ack_events
    WHERE command_id = c.command_id
    ORDER BY occurred_at DESC, event_id DESC
    LIMIT 1
) a ON true
WHERE c.tenant_id = $1
  AND c.product_key = $2
  AND c.device_slug = $3
  AND ($4 = '' OR c.service_name = $4)
ORDER BY c.occurred_at DESC, c.command_id DESC
LIMIT $6 OFFSET $7`

	var (
		tenantID   string
		productID  string
		productKey string
		deviceSlug string
		calls      []service.DeviceServiceCallHistoryEntry
		total      int
	)
	err := s.withDeviceUser(ctx, in.UserID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, scopeQuery, in.DeviceID).Scan(
			&tenantID,
			&productID,
			&productKey,
			&deviceSlug,
		); err != nil {
			return err
		}

		if err := tx.QueryRow(ctx, countQuery, tenantID, productKey, deviceSlug, in.ServiceName).Scan(&total); err != nil {
			return err
		}

		rows, err := tx.Query(
			ctx,
			query,
			tenantID,
			productKey,
			deviceSlug,
			in.ServiceName,
			in.AckDeadlineSeconds,
			in.PageInput.Limit(),
			in.PageInput.Offset(),
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var entry service.DeviceServiceCallHistoryEntry
			var input []byte
			var ackOutput []byte
			var ackSuccess sql.NullBool
			var ackOccurredAt sql.NullTime
			var ackReceivedAt sql.NullTime
			if err := rows.Scan(
				&entry.CommandID,
				&entry.TenantID,
				&entry.ProductKey,
				&entry.DeviceSlug,
				&entry.ServiceName,
				&entry.Topic,
				&input,
				&entry.OccurredAt,
				&entry.AckDeadlineAt,
				&entry.AckStatus,
				&entry.AckEventID,
				&ackSuccess,
				&entry.AckCode,
				&entry.AckMessage,
				&ackOutput,
				&ackOccurredAt,
				&ackReceivedAt,
			); err != nil {
				return err
			}
			entry.ProductID = productID
			entry.AckDeadlineSeconds = in.AckDeadlineSeconds
			entry.Input = map[string]any{}
			if len(input) > 0 {
				if err := json.Unmarshal(input, &entry.Input); err != nil {
					return fmt.Errorf("unmarshal device service call input: %w", err)
				}
			}
			entry.AckOutput = map[string]any{}
			if len(ackOutput) > 0 {
				if err := json.Unmarshal(ackOutput, &entry.AckOutput); err != nil {
					return fmt.Errorf("unmarshal device service call ack output: %w", err)
				}
			}
			if ackSuccess.Valid {
				success := ackSuccess.Bool
				entry.AckSuccess = &success
			}
			if ackOccurredAt.Valid {
				t := ackOccurredAt.Time
				entry.AckOccurredAt = &t
			}
			if ackReceivedAt.Valid {
				t := ackReceivedAt.Time
				entry.AckReceivedAt = &t
			}
			calls = append(calls, entry)
		}
		return rows.Err()
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.PageResult[service.DeviceServiceCallHistoryEntry]{}, service.ErrDeviceNotFound
		}
		return service.PageResult[service.DeviceServiceCallHistoryEntry]{}, fmt.Errorf("list device service call history: %w", err)
	}

	return service.NewPageResult(calls, total, in.PageInput), nil
}

func (s *PostgresDeviceStore) ListDevicePropertySetHistory(ctx context.Context, in service.DevicePropertySetHistoryInput) (service.PageResult[service.DevicePropertySetHistoryEntry], error) {
	if err := ctx.Err(); err != nil {
		return service.PageResult[service.DevicePropertySetHistoryEntry]{}, err
	}

	const scopeQuery = `
SELECT
    d.tenant_id::text,
    d.product_id::text,
    p.product_key,
    d.device_slug
FROM devices d
JOIN products p ON p.id = d.product_id AND p.tenant_id = d.tenant_id
WHERE d.id = $1`

	const countQuery = `
SELECT count(*)
FROM device_property_set_events
WHERE tenant_id = $1
  AND product_key = $2
  AND device_slug = $3
  AND ($4 = '' OR properties ? $4)`

	const query = `
SELECT
    c.command_id::text,
    c.tenant_id,
    c.product_key,
    c.device_slug,
    c.topic,
    c.properties,
    c.occurred_at,
    c.occurred_at + make_interval(secs => $5::int),
    CASE
      WHEN a.event_id IS NULL AND now() <= c.occurred_at + make_interval(secs => $5::int) THEN 'pending'
      WHEN a.event_id IS NULL THEN 'failed'
      WHEN a.success THEN 'success'
      ELSE 'failed'
    END,
    COALESCE(a.event_id::text, ''),
    a.success,
    COALESCE(a.code, ''),
    COALESCE(a.message, ''),
    COALESCE(a.properties, '{}'::jsonb),
    a.occurred_at,
    a.received_at
FROM device_property_set_events c
LEFT JOIN LATERAL (
    SELECT event_id, success, code, message, properties, occurred_at, received_at
    FROM device_property_set_ack_events
    WHERE command_id = c.command_id
    ORDER BY occurred_at DESC, event_id DESC
    LIMIT 1
) a ON true
WHERE c.tenant_id = $1
  AND c.product_key = $2
  AND c.device_slug = $3
  AND ($4 = '' OR c.properties ? $4)
ORDER BY c.occurred_at DESC, c.command_id DESC
LIMIT $6 OFFSET $7`

	var (
		tenantID   string
		productID  string
		productKey string
		deviceSlug string
		sets       []service.DevicePropertySetHistoryEntry
		total      int
	)
	err := s.withDeviceUser(ctx, in.UserID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, scopeQuery, in.DeviceID).Scan(
			&tenantID,
			&productID,
			&productKey,
			&deviceSlug,
		); err != nil {
			return err
		}

		if err := tx.QueryRow(ctx, countQuery, tenantID, productKey, deviceSlug, in.PropertyName).Scan(&total); err != nil {
			return err
		}

		rows, err := tx.Query(
			ctx,
			query,
			tenantID,
			productKey,
			deviceSlug,
			in.PropertyName,
			in.AckDeadlineSeconds,
			in.PageInput.Limit(),
			in.PageInput.Offset(),
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var entry service.DevicePropertySetHistoryEntry
			var properties []byte
			var ackProperties []byte
			var ackSuccess sql.NullBool
			var ackOccurredAt sql.NullTime
			var ackReceivedAt sql.NullTime
			if err := rows.Scan(
				&entry.CommandID,
				&entry.TenantID,
				&entry.ProductKey,
				&entry.DeviceSlug,
				&entry.Topic,
				&properties,
				&entry.OccurredAt,
				&entry.AckDeadlineAt,
				&entry.AckStatus,
				&entry.AckEventID,
				&ackSuccess,
				&entry.AckCode,
				&entry.AckMessage,
				&ackProperties,
				&ackOccurredAt,
				&ackReceivedAt,
			); err != nil {
				return err
			}
			entry.ProductID = productID
			entry.AckDeadlineSeconds = in.AckDeadlineSeconds
			entry.Properties = map[string]any{}
			if len(properties) > 0 {
				if err := json.Unmarshal(properties, &entry.Properties); err != nil {
					return fmt.Errorf("unmarshal device property set properties: %w", err)
				}
			}
			entry.AckProperties = map[string]any{}
			if len(ackProperties) > 0 {
				if err := json.Unmarshal(ackProperties, &entry.AckProperties); err != nil {
					return fmt.Errorf("unmarshal device property set ack properties: %w", err)
				}
			}
			if ackSuccess.Valid {
				success := ackSuccess.Bool
				entry.AckSuccess = &success
			}
			if ackOccurredAt.Valid {
				t := ackOccurredAt.Time
				entry.AckOccurredAt = &t
			}
			if ackReceivedAt.Valid {
				t := ackReceivedAt.Time
				entry.AckReceivedAt = &t
			}
			sets = append(sets, entry)
		}
		return rows.Err()
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.PageResult[service.DevicePropertySetHistoryEntry]{}, service.ErrDeviceNotFound
		}
		return service.PageResult[service.DevicePropertySetHistoryEntry]{}, fmt.Errorf("list device property set history: %w", err)
	}

	return service.NewPageResult(sets, total, in.PageInput), nil
}

func (s *PostgresDeviceStore) FindDeviceServiceCallTarget(ctx context.Context, in service.DeviceServiceCallTargetInput) (service.DeviceServiceCallTarget, error) {
	if err := ctx.Err(); err != nil {
		return service.DeviceServiceCallTarget{}, err
	}

	const query = `
	SELECT
	    t.id::text,
	    p.id::text,
	    d.id::text,
	    t.tenant_slug,
	    p.product_key,
	    d.device_slug,
	    d.status,
	    d.connection_status,
	    COALESCE(tm.services, '{}'::jsonb)
	FROM devices d
	JOIN tenants t ON t.id = d.tenant_id
	JOIN products p ON p.id = d.product_id AND p.tenant_id = d.tenant_id
	LEFT JOIN LATERAL (
	    SELECT services
	    FROM thingsmodel
	    WHERE tenant_id = d.tenant_id
	      AND product_id = d.product_id
	      AND is_current = true
	    LIMIT 1
	) tm ON true
	WHERE d.id = $1`

	var target service.DeviceServiceCallTarget
	var servicesRaw []byte
	err := s.withDeviceUser(ctx, in.UserID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query, in.DeviceID).Scan(
			&target.TenantID,
			&target.ProductID,
			&target.DeviceID,
			&target.TenantSlug,
			&target.ProductKey,
			&target.DeviceSlug,
			&target.DeviceStatus,
			&target.ConnectionStatus,
			&servicesRaw,
		)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.DeviceServiceCallTarget{}, service.ErrDeviceNotFound
		}
		return service.DeviceServiceCallTarget{}, fmt.Errorf("find device service call target: %w", err)
	}
	if err := json.Unmarshal(servicesRaw, &target.Services); err != nil {
		return service.DeviceServiceCallTarget{}, fmt.Errorf("decode device service call services: %w", err)
	}
	if target.Services == nil {
		target.Services = service.ThingsModelObject{}
	}
	return target, nil
}

func (s *PostgresDeviceStore) FindDevicePropertySetTarget(ctx context.Context, in service.DevicePropertySetTargetInput) (service.DevicePropertySetTarget, error) {
	if err := ctx.Err(); err != nil {
		return service.DevicePropertySetTarget{}, err
	}

	const query = `
	SELECT
	    t.id::text,
	    p.id::text,
	    d.id::text,
	    t.tenant_slug,
	    p.product_key,
	    d.device_slug,
	    d.status,
	    d.connection_status,
	    COALESCE(tm.properties, '{}'::jsonb)
	FROM devices d
	JOIN tenants t ON t.id = d.tenant_id
	JOIN products p ON p.id = d.product_id AND p.tenant_id = d.tenant_id
	LEFT JOIN LATERAL (
	    SELECT properties
	    FROM thingsmodel
	    WHERE tenant_id = d.tenant_id
	      AND product_id = d.product_id
	      AND is_current = true
	    LIMIT 1
	) tm ON true
	WHERE d.id = $1`

	var target service.DevicePropertySetTarget
	var propertiesRaw []byte
	err := s.withDeviceUser(ctx, in.UserID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query, in.DeviceID).Scan(
			&target.TenantID,
			&target.ProductID,
			&target.DeviceID,
			&target.TenantSlug,
			&target.ProductKey,
			&target.DeviceSlug,
			&target.DeviceStatus,
			&target.ConnectionStatus,
			&propertiesRaw,
		)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.DevicePropertySetTarget{}, service.ErrDeviceNotFound
		}
		return service.DevicePropertySetTarget{}, fmt.Errorf("find device property set target: %w", err)
	}
	if err := json.Unmarshal(propertiesRaw, &target.Properties); err != nil {
		return service.DevicePropertySetTarget{}, fmt.Errorf("decode device property set properties: %w", err)
	}
	if target.Properties == nil {
		target.Properties = service.ThingsModelObject{}
	}
	return target, nil
}

func (s *PostgresDeviceStore) UpdateDevice(ctx context.Context, in service.DeviceUpdateInput) (service.Device, error) {
	if err := ctx.Err(); err != nil {
		return service.Device{}, err
	}

	const query = `
UPDATE devices
SET device_name = $2,
    description = $3,
    status = $4,
    gateway_device_id = $5,
    updated_at = now()
WHERE id = $1
	RETURNING id::text, tenant_id::text, product_id::text, device_slug, device_name, description, status, connection_status, COALESCE(gateway_device_id::text, ''), COALESCE(firmware_version, ''), COALESCE(ip_address, ''), last_seen_at, created_at, updated_at`

	var device service.Device
	err := s.withDeviceUser(ctx, in.UserID, func(tx pgx.Tx) error {
		if in.GatewayDeviceID != "" {
			ok, err := gatewayBelongsToDeviceTenant(ctx, tx, in.GatewayDeviceID, in.DeviceID)
			if err != nil {
				return err
			}
			if !ok {
				return service.ErrInvalidDeviceInput
			}
		}

		return scanDevice(tx.QueryRow(
			ctx,
			query,
			in.DeviceID,
			in.DeviceName,
			in.Description,
			in.Status,
			nullIfEmpty(in.GatewayDeviceID),
		), &device)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.Device{}, service.ErrDeviceNotFound
		}
		return service.Device{}, fmt.Errorf("update device: %w", err)
	}
	return device, nil
}

func (s *PostgresDeviceStore) DeleteDevice(ctx context.Context, in service.DeviceDeleteInput) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	const query = `DELETE FROM devices WHERE id = $1`

	err := s.withDeviceUser(ctx, in.UserID, func(tx pgx.Tx) error {
		inUse, err := deviceHasSubDevices(ctx, tx, in.DeviceID)
		if err != nil {
			return err
		}
		if inUse {
			return service.ErrInvalidDeviceInput
		}

		tag, err := tx.Exec(ctx, query, in.DeviceID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return service.ErrDeviceNotFound
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, service.ErrDeviceNotFound) {
			return service.ErrDeviceNotFound
		}
		return fmt.Errorf("delete device: %w", err)
	}
	return nil
}

func (s *PostgresDeviceStore) withDeviceUser(ctx context.Context, userID string, fn func(pgx.Tx) error) error {
	return s.actor.withActor(ctx, userID, fn)
}

type deviceScanner interface {
	Scan(dest ...any) error
}

func scanDevice(row deviceScanner, device *service.Device) error {
	var lastSeenAt sql.NullTime
	if err := row.Scan(
		&device.ID,
		&device.TenantID,
		&device.ProductID,
		&device.DeviceSlug,
		&device.DeviceName,
		&device.Description,
		&device.Status,
		&device.ConnectionStatus,
		&device.GatewayDeviceID,
		&device.FirmwareVersion,
		&device.IPAddress,
		&lastSeenAt,
		&device.CreatedAt,
		&device.UpdatedAt,
	); err != nil {
		return err
	}
	if lastSeenAt.Valid {
		t := lastSeenAt.Time
		device.LastSeenAt = &t
	}
	return nil
}

func productBelongsToTenant(ctx context.Context, tx pgx.Tx, productID string, tenantID string) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM products WHERE id = $1 AND tenant_id = $2)`

	var ok bool
	if err := tx.QueryRow(ctx, query, productID, tenantID).Scan(&ok); err != nil {
		return false, fmt.Errorf("check device product: %w", err)
	}
	return ok, nil
}

func deviceBelongsToTenant(ctx context.Context, tx pgx.Tx, deviceID string, tenantID string) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM devices WHERE id = $1 AND tenant_id = $2)`

	var ok bool
	if err := tx.QueryRow(ctx, query, deviceID, tenantID).Scan(&ok); err != nil {
		return false, fmt.Errorf("check tenant device: %w", err)
	}
	return ok, nil
}

func gatewayBelongsToDeviceTenant(ctx context.Context, tx pgx.Tx, gatewayDeviceID string, deviceID string) (bool, error) {
	const query = `
SELECT EXISTS (
    SELECT 1
    FROM devices device
    JOIN devices gateway ON gateway.tenant_id = device.tenant_id
    WHERE device.id = $1
      AND gateway.id = $2
)`

	var ok bool
	if err := tx.QueryRow(ctx, query, deviceID, gatewayDeviceID).Scan(&ok); err != nil {
		return false, fmt.Errorf("check device gateway: %w", err)
	}
	return ok, nil
}

func deviceHasSubDevices(ctx context.Context, tx pgx.Tx, deviceID string) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM devices WHERE gateway_device_id = $1)`

	var ok bool
	if err := tx.QueryRow(ctx, query, deviceID).Scan(&ok); err != nil {
		return false, fmt.Errorf("check sub devices: %w", err)
	}
	return ok, nil
}

func insertActiveDeviceCredential(ctx context.Context, tx pgx.Tx, tenantID string, deviceID string, secretHash string) error {
	const query = `
INSERT INTO device_credentials (
    tenant_id,
    device_id,
    secret_hash,
    status
) VALUES (
    $1, $2, $3, 'active'
)`

	if _, err := tx.Exec(ctx, query, tenantID, deviceID, secretHash); err != nil {
		return fmt.Errorf("insert device credential: %w", err)
	}
	return nil
}
