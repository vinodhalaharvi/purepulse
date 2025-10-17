-- ============================================================================
-- EMPLOYEE ACTIVITY TRACKER - INITIAL SCHEMA
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

-- ============================================================================
-- TABLE: users
-- ============================================================================

CREATE TABLE users (
    user_id             TEXT PRIMARY KEY,
    display_name        TEXT,
    email               TEXT,
    timezone            TEXT DEFAULT 'UTC',
    active              BOOLEAN DEFAULT true,
    last_activity       TIMESTAMPTZ,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_last_activity ON users (last_activity DESC);

-- ============================================================================
-- TABLE: teams
-- ============================================================================

CREATE TABLE teams (
    team_id             TEXT PRIMARY KEY,
    team_name           TEXT NOT NULL,
    description         TEXT,
    parent_team_id      TEXT,
    metadata            JSONB DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

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
    PRIMARY KEY (team_id, user_id, joined_at),
    CONSTRAINT fk_team_members_team
        FOREIGN KEY (team_id)
        REFERENCES teams(team_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_team_members_user
        FOREIGN KEY (user_id)
        REFERENCES users(user_id)
        ON DELETE CASCADE
);

CREATE INDEX idx_team_members_user ON team_members (user_id) WHERE left_at IS NULL;
CREATE INDEX idx_team_members_team ON team_members (team_id) WHERE left_at IS NULL;

-- ============================================================================
-- TABLE: events
-- ============================================================================

CREATE TABLE events (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    event_uuid UUID NOT NULL DEFAULT uuid_generate_v4(),
    user_id TEXT NOT NULL,
    source platform_type NOT NULL,
    type event_type NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    payload JSONB NOT NULL,
    author TEXT,
    channel TEXT,
    thread_id TEXT,
    parent_id TEXT,
    size INTEGER,
    duration_seconds INTEGER,
    participants TEXT[],
    tags TEXT[],
    related_event_ids TEXT[],
    ingested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    search_vector tsvector GENERATED ALWAYS AS (
        to_tsvector('english',
            COALESCE(author, '') || ' ' ||
            COALESCE(channel, '') || ' ' ||
            COALESCE(payload::text, '')
        )
    ) STORED,
    CONSTRAINT fk_events_user
        FOREIGN KEY (user_id)
        REFERENCES users(user_id)
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX events_event_uuid_key ON events(event_uuid);
CREATE INDEX idx_events_user_time ON events(user_id, timestamp DESC);
CREATE INDEX idx_events_source_time ON events(source, timestamp DESC);
CREATE INDEX idx_events_type_time ON events (type, timestamp DESC);
CREATE INDEX idx_events_user_source ON events (user_id, source);
CREATE INDEX idx_events_channel ON events (channel) WHERE channel IS NOT NULL;
CREATE INDEX idx_events_author ON events (author) WHERE author IS NOT NULL;
CREATE INDEX idx_events_payload_gin ON events USING GIN (payload jsonb_path_ops);
CREATE INDEX idx_events_search ON events USING GIN (search_vector);
CREATE INDEX idx_events_user_source_time ON events (user_id, source, timestamp DESC);

-- ============================================================================
-- TABLE: weekly_reports
-- ============================================================================

-- ============================================================================
-- TABLE: weekly_reports (User reports)
-- ============================================================================

CREATE TABLE weekly_reports (
                                id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                user_id                 TEXT NOT NULL,
                                week_start              TIMESTAMPTZ NOT NULL,
                                week_end                TIMESTAMPTZ NOT NULL,
                                executive_summary       TEXT,
                                wins                    JSONB NOT NULL DEFAULT '[]'::jsonb,
                                in_progress             JSONB NOT NULL DEFAULT '[]'::jsonb,
                                blocked                 JSONB NOT NULL DEFAULT '[]'::jsonb,
                                recommendations         TEXT[] DEFAULT '{}',
                                ai_model                TEXT,
                                ai_tokens               INTEGER,
                                ai_latency_ms           INTEGER,
                                generated_at            TIMESTAMPTZ DEFAULT NOW(),
                                version                 TEXT DEFAULT '1.0',
                                CONSTRAINT fk_weekly_reports_user
                                    FOREIGN KEY (user_id)
        REFERENCES users(user_id)
        ON DELETE CASCADE,
                                UNIQUE (user_id, week_start, week_end)
);

CREATE INDEX idx_weekly_reports_user_week ON weekly_reports (user_id, week_start DESC);
CREATE INDEX idx_weekly_reports_week ON weekly_reports (week_start DESC);

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
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- SEED: Sample Data
-- ============================================================================

INSERT INTO users (user_id, display_name, email, active)
VALUES
    ('alice', 'Alice User', 'alice@example.com', true),
    ('bob', 'Bob User', 'bob@example.com', true),
    ('charlie', 'Charlie User', 'charlie@example.com', true)
ON CONFLICT (user_id) DO NOTHING;

INSERT INTO teams (team_id, team_name, description)
VALUES
    ('team-engineering', 'Engineering Team', 'Core platform engineering')
ON CONFLICT (team_id) DO NOTHING;

INSERT INTO team_members (team_id, user_id, role)
VALUES
    ('team-engineering', 'alice', 'senior-engineer'),
    ('team-engineering', 'bob', 'engineer'),
    ('team-engineering', 'charlie', 'engineer')
ON CONFLICT DO NOTHING;





CREATE TABLE team_weekly_reports (
                                     id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                     team_id                 TEXT NOT NULL,
                                     week_start              TIMESTAMPTZ NOT NULL,
                                     week_end                TIMESTAMPTZ NOT NULL,
                                     executive_summary       TEXT,
                                     velocity_analysis       TEXT,
                                     collaboration_notes     TEXT,
                                     team_blockers           JSONB DEFAULT '[]'::jsonb,
                                     recommendations         TEXT[] DEFAULT '{}',
                                     ai_model                TEXT,
                                     ai_tokens               INTEGER,
                                     ai_latency_ms           INTEGER,
                                     generated_at            TIMESTAMPTZ DEFAULT NOW(),
                                     version                 TEXT DEFAULT '1.0',
                                     CONSTRAINT fk_team_weekly_reports_team
                                         FOREIGN KEY (team_id)
        REFERENCES teams(team_id)
        ON DELETE CASCADE,
                                     UNIQUE (team_id, week_start, week_end)
);

CREATE INDEX idx_team_weekly_reports_team_week ON team_weekly_reports (team_id, week_start DESC);
CREATE INDEX idx_team_weekly_reports_week ON team_weekly_reports (week_start DESC);