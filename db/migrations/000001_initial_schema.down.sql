-- ============================================================================
-- ROLLBACK: Drop entire schema
-- ============================================================================

DROP TRIGGER IF EXISTS update_events_updated_at ON events;
DROP TRIGGER IF EXISTS update_teams_updated_at ON teams;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

DROP FUNCTION IF EXISTS refresh_all_materialized_views();
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP MATERIALIZED VIEW IF EXISTS daily_user_activity CASCADE;

DROP TABLE IF EXISTS weekly_reports CASCADE;
DROP TABLE IF EXISTS events CASCADE;
DROP TABLE IF EXISTS team_members CASCADE;
DROP TABLE IF EXISTS teams CASCADE;
DROP TABLE IF EXISTS users CASCADE;

DROP TYPE IF EXISTS event_type;
DROP TYPE IF EXISTS platform_type;

DROP EXTENSION IF EXISTS "pg_stat_statements";
DROP EXTENSION IF EXISTS "btree_gin";
DROP EXTENSION IF EXISTS "pg_trgm";
DROP EXTENSION IF EXISTS "uuid-ossp";


DROP TABLE IF EXISTS team_weekly_reports CASCADE;
DROP TABLE IF EXISTS weekly_reports CASCADE;