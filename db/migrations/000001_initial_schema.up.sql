-- ============================================================================
-- EMPLOYEE ACTIVITY TRACKER - INITIAL SCHEMA (FIXED)
-- ============================================================================

-- Enable extensions first
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gin";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

-- ============================================================================
-- ENUMS
-- ============================================================================

CREATE TYPE platform_type AS ENUM (
    'slack',
    'github',
    'jira',
    'zoom'
);

CREATE TYPE event_type AS ENUM (
    'slack_message',
    'slack_reaction',
    'slack_file_upload',
    'slack_channel_join',
    'slack_thread_reply',
    'github_commit',
    'github_pull_request',
    'github_pr_review',
    'github_pr_comment',
    'github_issue',
    'github_issue_comment',
    'github_release',
    'jira_issue_created',
    'jira_issue_updated',
    'jira_issue_closed',
    'jira_comment',
    'jira_transition',
    'jira_sprint',
    'zoom_meeting',
    'zoom_webinar',
    'zoom_recording',
    'zoom_chat'
);

CREATE TYPE correlation_type AS ENUM (
    'slack_to_github',
    'jira_to_github',
    'zoom_to_activity',
    'github_to_jira',
    'slack_to_jira'
);

-- ============================================================================
-- TABLE: users (must come before events due to FK)
-- ============================================================================

CREATE TABLE users (
    user_id             TEXT PRIMARY KEY,
    display_name        TEXT,
    email               TEXT,
    timezone            TEXT DEFAULT 'UTC',
    slack_id            TEXT,
    github_login        TEXT,
    jira_account_id     TEXT,
    zoom_email          TEXT,
    platform_identities JSONB,
    active              BOOLEAN DEFAULT true,
    last_activity       TIMESTAMPTZ,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_slack_id ON users (slack_id) WHERE slack_id IS NOT NULL;
CREATE INDEX idx_users_github_login ON users (github_login) WHERE github_login IS NOT NULL;
CREATE INDEX idx_users_jira_account_id ON users (jira_account_id) WHERE jira_account_id IS NOT NULL;
CREATE INDEX idx_users_last_activity ON users (last_activity DESC);

-- ============================================================================
-- TABLE: teams
-- ============================================================================

CREATE TABLE teams (
    team_id             TEXT PRIMARY KEY,
    team_name           TEXT NOT NULL,
    description         TEXT,
    parent_team_id      TEXT,  -- FK added later to avoid circular reference
    metadata            JSONB DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- Add self-referencing FK after table creation
ALTER TABLE teams
ADD CONSTRAINT fk_teams_parent
FOREIGN KEY (parent_team_id)
REFERENCES teams(team_id);

CREATE INDEX idx_teams_parent ON teams (parent_team_id) WHERE parent_team_id IS NOT NULL;

-- ============================================================================
-- TABLE: team_members
-- ============================================================================

CREATE TABLE team_members (
    team_id             TEXT NOT NULL,
    user_id             TEXT NOT NULL,
    role                TEXT DEFAULT 'member',
    joined_at           TIMESTAMPTZ DEFAULT NOW(),
    left_at             TIMESTAMPTZ,
    PRIMARY KEY (team_id, user_id, joined_at)
);

-- Add FKs after table creation
ALTER TABLE team_members
ADD CONSTRAINT fk_team_members_team
FOREIGN KEY (team_id)
REFERENCES teams(team_id)
ON DELETE CASCADE;

ALTER TABLE team_members
ADD CONSTRAINT fk_team_members_user
FOREIGN KEY (user_id)
REFERENCES users(user_id)
ON DELETE CASCADE;

CREATE INDEX idx_team_members_user ON team_members (user_id) WHERE left_at IS NULL;
CREATE INDEX idx_team_members_team ON team_members (team_id) WHERE left_at IS NULL;

-- ============================================================================
-- TABLE: events (partitioned)
-- ============================================================================

CREATE TABLE events (
    id                  BIGSERIAL,
    event_uuid          UUID DEFAULT uuid_generate_v4() NOT NULL,
    user_id             TEXT NOT NULL,
    source              platform_type NOT NULL,
    type                event_type NOT NULL,
    timestamp           TIMESTAMPTZ NOT NULL,
    payload             JSONB NOT NULL,
    author              TEXT,
    channel             TEXT,
    thread_id           TEXT,
    parent_id           TEXT,
    size                INTEGER,
    duration_seconds    INTEGER,
    participants        TEXT[],
    tags                TEXT[],
    related_event_ids   TEXT[],
    ingested_at         TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at          TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    search_vector       tsvector GENERATED ALWAYS AS (
        to_tsvector('english', 
            COALESCE(author, '') || ' ' ||
            COALESCE(channel, '') || ' ' ||
            COALESCE(payload::text, '')
        )
    ) STORED,
    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

-- Create partitions
DO $$
DECLARE
    start_date DATE := DATE_TRUNC('month', NOW());
    end_date DATE;
    partition_name TEXT;
BEGIN
    FOR i IN 0..6 LOOP
        end_date := start_date + INTERVAL '1 month';
        partition_name := 'events_' || TO_CHAR(start_date, 'YYYY_MM');

        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF events
             FOR VALUES FROM (%L) TO (%L)',
            partition_name,
            start_date,
            end_date
        );

        start_date := end_date;
    END LOOP;
END $$;

-- Indexes
CREATE INDEX idx_events_user_time ON events (user_id, timestamp DESC);
CREATE INDEX idx_events_source_time ON events (source, timestamp DESC);
CREATE INDEX idx_events_type_time ON events (type, timestamp DESC);
CREATE INDEX idx_events_user_source ON events (user_id, source);
CREATE INDEX idx_events_channel ON events (channel) WHERE channel IS NOT NULL;
CREATE INDEX idx_events_author ON events (author) WHERE author IS NOT NULL;
CREATE INDEX idx_events_payload_gin ON events USING GIN (payload jsonb_path_ops);
CREATE INDEX idx_events_search ON events USING GIN (search_vector);
CREATE INDEX idx_events_user_source_time ON events (user_id, source, timestamp DESC);

-- ============================================================================
-- TABLE: correlations
-- ============================================================================

CREATE TABLE correlations (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             TEXT NOT NULL,
    type                correlation_type NOT NULL,
    source_event_id     BIGINT NOT NULL,
    target_event_id     BIGINT NOT NULL,
    confidence          NUMERIC(3, 2) NOT NULL CHECK (confidence BETWEEN 0 AND 1),
    time_delta_seconds  INTEGER NOT NULL,
    frequency           INTEGER DEFAULT 1,
    description         TEXT,
    metadata            JSONB DEFAULT '{}'::jsonb,
    detected_at         TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (user_id, type, source_event_id, target_event_id)
);

-- Add FK after table creation
ALTER TABLE correlations
ADD CONSTRAINT fk_correlations_user
FOREIGN KEY (user_id)
REFERENCES users(user_id);

CREATE INDEX idx_correlations_user ON correlations (user_id, detected_at DESC);
CREATE INDEX idx_correlations_type ON correlations (type, confidence DESC);
CREATE INDEX idx_correlations_confidence ON correlations (confidence DESC) WHERE confidence >= 0.7;
CREATE INDEX idx_correlations_source_event ON correlations (source_event_id);
CREATE INDEX idx_correlations_target_event ON correlations (target_event_id);

-- ============================================================================
-- TABLE: summaries
-- ============================================================================

CREATE TABLE summaries (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             TEXT NOT NULL,
    time_range_start    TIMESTAMPTZ NOT NULL,
    time_range_end      TIMESTAMPTZ NOT NULL,
    activity            JSONB NOT NULL,
    metrics             JSONB NOT NULL,
    correlations        JSONB NOT NULL,
    ai_summary          JSONB,
    ai_model            TEXT,
    ai_tokens           INTEGER,
    ai_latency_ms       INTEGER,
    audit_log           TEXT[],
    version             TEXT DEFAULT '1.0',
    generated_at        TIMESTAMPTZ DEFAULT NOW(),
    expires_at          TIMESTAMPTZ,
    UNIQUE (user_id, time_range_start, time_range_end)
);

ALTER TABLE summaries
ADD CONSTRAINT fk_summaries_user
FOREIGN KEY (user_id)
REFERENCES users(user_id);

CREATE INDEX idx_summaries_user_time ON summaries (user_id, time_range_start DESC, time_range_end DESC);
CREATE INDEX idx_summaries_expires ON summaries (expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_summaries_generated ON summaries (generated_at DESC);

-- ============================================================================
-- TABLE: team_summaries
-- ============================================================================

CREATE TABLE team_summaries (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id             TEXT NOT NULL,
    time_range_start    TIMESTAMPTZ NOT NULL,
    time_range_end      TIMESTAMPTZ NOT NULL,
    member_summaries    JSONB NOT NULL,
    aggregated_metrics  JSONB NOT NULL,
    generated_at        TIMESTAMPTZ DEFAULT NOW(),
    expires_at          TIMESTAMPTZ,
    UNIQUE (team_id, time_range_start, time_range_end)
);

ALTER TABLE team_summaries
ADD CONSTRAINT fk_team_summaries_team
FOREIGN KEY (team_id)
REFERENCES teams(team_id);

CREATE INDEX idx_team_summaries_team_time ON team_summaries (team_id, time_range_start DESC);
CREATE INDEX idx_team_summaries_expires ON team_summaries (expires_at) WHERE expires_at IS NOT NULL;

-- ============================================================================
-- TABLE: fetch_metadata
-- ============================================================================

CREATE TABLE fetch_metadata (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             TEXT NOT NULL,
    source              platform_type NOT NULL,
    time_range_start    TIMESTAMPTZ NOT NULL,
    time_range_end      TIMESTAMPTZ NOT NULL,
    events_fetched      INTEGER DEFAULT 0,
    errors              JSONB,
    api_calls           INTEGER DEFAULT 0,
    latency_ms          INTEGER,
    cached              BOOLEAN DEFAULT false,
    rate_limited        BOOLEAN DEFAULT false,
    fetched_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_fetch_metadata_user_source ON fetch_metadata (user_id, source, fetched_at DESC);
CREATE INDEX idx_fetch_metadata_errors ON fetch_metadata (source, fetched_at DESC) WHERE errors IS NOT NULL;

-- ============================================================================
-- TABLE: audit_log
-- ============================================================================

CREATE TABLE audit_log (
    id                  BIGSERIAL PRIMARY KEY,
    operation           TEXT NOT NULL,
    user_id             TEXT,
    table_name          TEXT,
    record_id           TEXT,
    changes             JSONB,
    metadata            JSONB DEFAULT '{}'::jsonb,
    request_id          UUID,
    ip_address          INET,
    user_agent          TEXT,
    timestamp           TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_audit_log_timestamp ON audit_log (timestamp DESC);
CREATE INDEX idx_audit_log_operation ON audit_log (operation, timestamp DESC);
CREATE INDEX idx_audit_log_user ON audit_log (user_id, timestamp DESC) WHERE user_id IS NOT NULL;
CREATE INDEX idx_audit_log_request ON audit_log (request_id) WHERE request_id IS NOT NULL;

-- ============================================================================
-- TABLE: api_keys
-- ============================================================================

CREATE TABLE api_keys (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    platform            platform_type NOT NULL,
    key_name            TEXT NOT NULL,
    encrypted_token     TEXT NOT NULL,
    encryption_key_id   TEXT NOT NULL,
    workspace_id        TEXT,
    scopes              TEXT[],
    active              BOOLEAN DEFAULT true,
    rate_limit_config   JSONB,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    created_by          TEXT,
    last_used_at        TIMESTAMPTZ,
    expires_at          TIMESTAMPTZ,
    UNIQUE (platform, key_name)
);

CREATE INDEX idx_api_keys_platform ON api_keys (platform) WHERE active = true;
CREATE INDEX idx_api_keys_expires ON api_keys (expires_at) WHERE expires_at IS NOT NULL;

-- ============================================================================
-- MATERIALIZED VIEWS
-- ============================================================================

CREATE MATERIALIZED VIEW daily_user_activity AS
SELECT
    user_id,
    DATE(timestamp) AS date,
    source,
    type,
    COUNT(*) AS event_count,
    MIN(timestamp) AS first_event_at,
    MAX(timestamp) AS last_event_at,
    ARRAY_AGG(DISTINCT channel) FILTER (WHERE channel IS NOT NULL) AS channels,
    ARRAY_AGG(DISTINCT author) FILTER (WHERE author IS NOT NULL) AS collaborators,
    SUM(size) FILTER (WHERE size IS NOT NULL) AS total_size,
    SUM(duration_seconds) FILTER (WHERE duration_seconds IS NOT NULL) AS total_duration_seconds
FROM events
GROUP BY user_id, DATE(timestamp), source, type;

CREATE UNIQUE INDEX idx_daily_user_activity_pk ON daily_user_activity (user_id, date, source, type);
CREATE INDEX idx_daily_user_activity_date ON daily_user_activity (date DESC);
CREATE INDEX idx_daily_user_activity_user ON daily_user_activity (user_id, date DESC);

CREATE MATERIALIZED VIEW weekly_team_activity AS
SELECT
    tm.team_id,
    DATE_TRUNC('week', e.timestamp) AS week_start,
    e.source,
    COUNT(DISTINCT tm.user_id) AS active_members,
    COUNT(*) AS total_events,
    SUM(e.size) FILTER (WHERE e.size IS NOT NULL) AS total_size,
    AVG(e.size) FILTER (WHERE e.size IS NOT NULL) AS avg_size
FROM events e
JOIN team_members tm ON e.user_id = tm.user_id
WHERE tm.left_at IS NULL
GROUP BY tm.team_id, DATE_TRUNC('week', e.timestamp), e.source;

CREATE UNIQUE INDEX idx_weekly_team_activity_pk ON weekly_team_activity (team_id, week_start, source);
CREATE INDEX idx_weekly_team_activity_week ON weekly_team_activity (week_start DESC);

CREATE MATERIALIZED VIEW user_correlation_summary AS
SELECT
    user_id,
    type,
    COUNT(*) AS pattern_count,
    AVG(confidence) AS avg_confidence,
    AVG(time_delta_seconds) AS avg_time_delta_seconds,
    SUM(frequency) AS total_frequency,
    MAX(detected_at) AS last_detected_at
FROM correlations
WHERE confidence >= 0.5
GROUP BY user_id, type;

CREATE UNIQUE INDEX idx_user_correlation_summary_pk ON user_correlation_summary (user_id, type);
CREATE INDEX idx_user_correlation_summary_user ON user_correlation_summary (user_id);

-- ============================================================================
-- FUNCTIONS & TRIGGERS
-- ============================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_teams_updated_at BEFORE UPDATE ON teams
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_events_updated_at BEFORE UPDATE ON events
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE OR REPLACE FUNCTION refresh_all_materialized_views()
    RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY daily_user_activity;
REFRESH MATERIALIZED VIEW CONCURRENTLY weekly_team_activity;
REFRESH MATERIALIZED VIEW CONCURRENTLY user_correlation_summary;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION cleanup_expired_summaries()
    RETURNS void AS $$
BEGIN
    DELETE FROM summaries WHERE expires_at < NOW();
DELETE FROM team_summaries WHERE expires_at < NOW();
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION create_next_month_partition()
    RETURNS void AS $$
DECLARE
    next_month_start DATE := DATE_TRUNC('month', NOW() + INTERVAL '1 month');
next_month_end DATE := next_month_start + INTERVAL '1 month';
partition_name TEXT := 'events_' || TO_CHAR(next_month_start, 'YYYY_MM');
BEGIN
    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF events
         FOR VALUES FROM (%L) TO (%L)',
        partition_name,
        next_month_start,
        next_month_end
    );
END;
$$ LANGUAGE plpgsql;