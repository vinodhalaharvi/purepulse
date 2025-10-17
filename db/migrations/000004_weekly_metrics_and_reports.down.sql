-- ============================================================================
-- ROLLBACK: Drop weekly metrics and reports tables
-- ============================================================================

DROP TRIGGER IF EXISTS update_team_weekly_metrics_updated_at ON team_weekly_metrics;
DROP TRIGGER IF EXISTS update_weekly_metrics_updated_at ON weekly_metrics;

DROP TABLE IF EXISTS weekly_reports CASCADE;
DROP TABLE IF EXISTS team_weekly_metrics CASCADE;
DROP TABLE IF EXISTS weekly_metrics CASCADE;

