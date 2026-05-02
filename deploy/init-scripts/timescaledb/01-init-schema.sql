-- ============================================================
-- LinkFlow v2 - Database bootstrap
-- ============================================================
-- This script runs ONCE when the TimescaleDB container starts with
-- an empty data volume. It only enables required extensions.
--
-- Business tables are intentionally NOT created here. Schema design
-- is driven by the event contracts and query needs of each service,
-- and will be added by the services themselves (via migrations) when
-- they are implemented.
--
-- To re-run this script during development:
--   make clean && make up
-- ============================================================

-- TimescaleDB: time-series capabilities (hypertables, retention, etc.)
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- uuid-ossp: uuid_generate_v4() for UUID primary keys
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- Bootstrap complete notice
-- ============================================================
DO $$
BEGIN
    RAISE NOTICE '====================================';
    RAISE NOTICE 'LinkFlow v2 database bootstrap complete';
    RAISE NOTICE 'PostgreSQL: %', current_setting('server_version');
    RAISE NOTICE 'TimescaleDB: %', (SELECT extversion FROM pg_extension WHERE extname = 'timescaledb');
    RAISE NOTICE 'Schemas will be added by services as needed.';
    RAISE NOTICE '====================================';
END $$;