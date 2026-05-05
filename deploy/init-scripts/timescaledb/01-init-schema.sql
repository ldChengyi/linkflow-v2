-- ============================================================
-- LinkFlow v2 - Database bootstrap
-- ============================================================
-- This script runs ONCE when the TimescaleDB container starts with
-- an empty data volume. It only enables required extensions.
--
-- Business tables are intentionally kept in separate files in this
-- directory. Docker runs these files in lexical order when the database
-- volume is first initialized.
--
-- To re-run this script during development:
--   scripts/db/apply-timescaledb.sh
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
    RAISE NOTICE 'Schema bootstrap files will run after extension initialization.';
    RAISE NOTICE '====================================';
END $$;
