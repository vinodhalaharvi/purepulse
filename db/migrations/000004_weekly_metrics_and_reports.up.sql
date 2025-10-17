-- ============================================================================
-- TABLE: weekly_metrics
-- ============================================================================

CREATE TABLE weekly_metrics (
    id                          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id                     TEXT NOT NULL,
    week_start                  TIMESTAMPTZ NOT NULL,
    week_end                    TIMESTAMPTZ NOT NULL,
    
    -- GitHub metrics
    commits                     INTEGER DEFAULT 0,
    commits_avg_quality         NUMERIC(3, 2) DEFAULT 0.0,
    prs_opened                  INTEGER DEFAULT 0,
    prs_reviewed                INTEGER DEFAULT 0,
    review_turnaround_ms        INTEGER DEFAULT 0,
    lines_added                 INTEGER DEFAULT 0,
    lines_deleted               INTEGER DEFAULT 0,
    
    -- Jira metrics
    tickets_completed           INTEGER DEFAULT 0,
    tickets_in_progress         INTEGER DEFAULT 0,
    tickets_blocked             INTEGER DEFAULT 0,
    avg_completion_days         NUMERIC(5, 2) DEFAULT 0.0,
    story_points_completed      NUMERIC(8, 2) DEFAULT 0.0,
    
    -- Slack metrics
    slack_messages              INTEGER DEFAULT 0,
    channels_active             TEXT[] DEFAULT '{}',
    pair_programming_hours      NUMERIC(5, 2) DEFAULT 0.0,
    help_questions_answered     INTEGER DEFAULT 0,
    
    -- Meeting metrics
    total_meeting_hours         NUMERIC(5, 2) DEFAULT 0.0,
    deep_work_hours             NUMERIC(5, 2) DEFAULT 0.0,
    context_switches            INTEGER DEFAULT 0,
    
    -- Patterns
    peak_activity_hours         INTEGER[] DEFAULT '{}',
    collaboration_score         NUMERIC(3, 2) DEFAULT 0.0,
    productivity_score          NUMERIC(3, 2) DEFAULT 0.0,
    
    -- Metadata
    generated_at                TIMESTAMPTZ DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT fk_weekly_metrics_user
        FOREIGN KEY (user_id)
        REFERENCES users(user_id)
        ON DELETE CASCADE,
    
    UNIQUE (user_id, week_start, week_end)
);

CREATE INDEX idx_weekly_metrics_user_week ON weekly_metrics (user_id, week_start DESC);
CREATE INDEX idx_weekly_metrics_week ON weekly_metrics (week_start DESC);

-- ============================================================================
-- TABLE: team_weekly_metrics
-- ============================================================================

CREATE TABLE team_weekly_metrics (
    id                          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id                     TEXT NOT NULL,
    week_start                  TIMESTAMPTZ NOT NULL,
    week_end                    TIMESTAMPTZ NOT NULL,
    
    -- Team aggregates
    total_velocity              INTEGER DEFAULT 0,
    velocity_trend              NUMERIC(5, 2) DEFAULT 0.0,
    blocker_count               INTEGER DEFAULT 0,
    avg_blocker_days            NUMERIC(5, 2) DEFAULT 0.0,
    team_collaboration_score    NUMERIC(3, 2) DEFAULT 0.0,
    
    -- Patterns
    top_collaborators           TEXT[] DEFAULT '{}',
    key_blockers                TEXT[] DEFAULT '{}',
    mentorship_pairs            JSONB DEFAULT '[]'::jsonb,
    
    -- Metadata
    generated_at                TIMESTAMPTZ DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT fk_team_weekly_metrics_team
        FOREIGN KEY (team_id)
        REFERENCES teams(team_id)
        ON DELETE CASCADE,
    
    UNIQUE (team_id, week_start, week_end)
);

CREATE INDEX idx_team_weekly_metrics_team_week ON team_weekly_metrics (team_id, week_start DESC);
CREATE INDEX idx_team_weekly_metrics_week ON team_weekly_metrics (week_start DESC);

-- ============================================================================
-- TABLE: weekly_reports
-- ============================================================================

CREATE TABLE weekly_reports (
    id                          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id                     TEXT NOT NULL,
    week_start                  TIMESTAMPTZ NOT NULL,
    week_end                    TIMESTAMPTZ NOT NULL,
    
    -- Report content (stored as JSONB for flexibility)
    executive_summary           TEXT,
    user_reports                JSONB NOT NULL,  -- map of user_id -> report
    velocity_analysis           TEXT,
    collaboration_notes         TEXT,
    team_blockers               JSONB DEFAULT '[]'::jsonb,
    recommendations             TEXT[] DEFAULT '{}',
    growth_opportunities        TEXT[] DEFAULT '{}',
    metrics_summary             JSONB DEFAULT '{}'::jsonb,
    
    -- LLM metadata
    ai_model                    TEXT,
    ai_tokens                   INTEGER,
    ai_latency_ms               INTEGER,
    
    -- Metadata
    generated_at                TIMESTAMPTZ DEFAULT NOW(),
    expires_at                  TIMESTAMPTZ,
    version                     TEXT DEFAULT '1.0',
    
    CONSTRAINT fk_weekly_reports_team
        FOREIGN KEY (team_id)
        REFERENCES teams(team_id)
        ON DELETE CASCADE,
    
    UNIQUE (team_id, week_start, week_end)
);

CREATE INDEX idx_weekly_reports_team_week ON weekly_reports (team_id, week_start DESC);
CREATE INDEX idx_weekly_reports_generated ON weekly_reports (generated_at DESC);
CREATE INDEX idx_weekly_reports_expires ON weekly_reports (expires_at) WHERE expires_at IS NOT NULL;

-- ============================================================================
-- TRIGGER: Update updated_at on weekly_metrics
-- ============================================================================

CREATE TRIGGER update_weekly_metrics_updated_at
BEFORE UPDATE ON weekly_metrics
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_team_weekly_metrics_updated_at
BEFORE UPDATE ON team_weekly_metrics
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();


