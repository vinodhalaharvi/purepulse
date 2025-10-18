# PurePulse Analytics Platform

A functional programming-based analytics platform that aggregates employee activity data from multiple sources (Slack,
GitHub, JIRA, Zoom) and generates AI-powered weekly reports using Claude.

## Architecture
```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Data Sources   │     │   PostgreSQL    │     │   GraphQL API   │
├─────────────────┤     ├─────────────────┤     ├─────────────────┤
│ • Slack         │────▶│ • Events        │────▶│ • User Weekly   │
│ • GitHub        │     │ • Users/Teams   │     │ • Team Weekly   │
│ • JIRA          │     │ • Reports       │     │ • Analytics     │
│ • Zoom          │     │ • Views         │     └─────────────────┘
└─────────────────┘     └─────────────────┘              ▲
                                 │                        │
                                 ▼                        │
                        ┌─────────────────┐     ┌─────────────────┐
                        │  Claude AI      │     │   Web Client    │
                        ├─────────────────┤     └─────────────────┘
                        │ Weekly Analysis │
                        │ Team Insights   │
                        └─────────────────┘
```

## Tech Stack

- **Language**: Go 1.21+
- **Database**: PostgreSQL 14+ with materialized views
- **API**: GraphQL (gqlgen)
- **AI**: Claude API (Anthropic)
- **Patterns**: Functional programming with monoids, functors, and effects

## Prerequisites

- Go 1.21 or higher
- Docker & Docker Compose
- PostgreSQL client (`psql`)
- Make
- Anthropic API key for Claude

## Quick Start

### 1. Clone and Setup

```bash
git clone https://github.com/vinodhalaharvi/purepulse.git
cd purepulse

# Copy environment template
cp .env.example .env

# Edit .env and add your Anthropic API key
# ANTHROPIC_API_KEY=your-key-here
```

### 2. Database Setup

```bash
# Start PostgreSQL container
make db-up

# Run migrations
make migrate-up

# Verify setup
make db-shell
\dt  # List tables
\q   # Exit
```

### 3. Generate Test Data and Reports

```bash
# Full stack setup (reset DB, ingest data, generate reports, start GraphQL)
make full-stack
```

This command will:

1. Reset database with fresh schema
2. Generate mock data for 5 users (alice, bob, charlie, david, emma)
3. Refresh materialized views
4. Generate user weekly reports with Claude
5. Generate team weekly report with Claude
6. Start GraphQL server at http://localhost:8080/

### 4. Query Data via GraphQL

Visit http://localhost:8080/ for GraphQL Playground

**Get User Weekly Report:**

```graphql
query {
    userWeekly(
        userID: "alice"
        weekStart: "2025-10-11"
        weekEnd: "2025-10-18"
    ) {
        userID
        wins {
            title
            description
            impact
        }
        inProgress {
            title
            percentComplete
        }
        blocked {
            title
            blockedBy
            suggestedAction
        }
        notes
        generatedAt
    }
}
```

**Get Team Weekly Report:**

```graphql
query {
    teamWeekly(
        teamID: "team-engineering"
        weekStart: "2025-10-11"
        weekEnd: "2025-10-18"
    ) {
        teamID
        executiveSummary
        velocityAnalysis
        collaborationNotes
        teamBlockers {
            title
            affectedUsers
            action
        }
        recommendations
        memberCount
    }
}
```

## Key Commands

### Database Management

```bash
make db-up          # Start PostgreSQL
make db-down        # Stop PostgreSQL
make db-reset       # Reset database (WARNING: deletes all data)
make db-shell       # Open psql shell
```

### Data Pipeline

```bash
make ingest              # Load mock data
make refresh-views       # Refresh materialized views
make test-user-weekly    # Generate user reports with Claude
make test-team-weekly    # Generate team report with Claude
```

### Development

```bash
make graphql-server      # Start GraphQL server
make full-stack          # Complete setup from scratch
go test ./...            # Run tests
```

### Check Data

```bash
# View events by user
psql $DATABASE_URL -c "SELECT user_id, COUNT(*) FROM events GROUP BY user_id;"

# View weekly reports
psql $DATABASE_URL -c "SELECT user_id, week_start FROM weekly_reports;"

# View team reports
psql $DATABASE_URL -c "SELECT team_id, week_start FROM team_weekly_reports;"
```

## Project Structure
```
purepulse/
├── cmd/
│   ├── purepulse/        # Data ingestion
│   ├── test-user-weekly/ # User report generation
│   └── test-team-weekly/ # Team report generation
├── db/
│   ├── migrations/        # SQL migrations
│   └── query/            # Query types
├── internal/
│   └── connectors/       # Platform connectors (Slack, GitHub, etc)
├── pkg/
│   ├── analytics/        # Analytics logic
│   ├── collectors/       # Data collection
│   ├── events/          # Event types
│   ├── graphql/         # GraphQL API
│   ├── llm/             # Claude AI integration
│   ├── monoids/         # Functional combinators
│   └── types/           # Shared types
└── Makefile             # Build commands
```

## Features

- **Multi-platform Integration**: Collects data from Slack, GitHub, JIRA, and Zoom
- **AI-Powered Analysis**: Uses Claude to generate insights and recommendations
- **Functional Architecture**: Built with monoids, functors, and effect types
- **Weekly Reports**: Automated individual and team performance summaries
- **GraphQL API**: Flexible querying of analytics data
- **Materialized Views**: Optimized aggregations for fast queries

## Environment Variables

Create a `.env` file with:

```env
# Database
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/purepulse?sslmode=disable
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=purepulse

# Claude AI
ANTHROPIC_API_KEY=your-api-key-here

# Server
PORT=8080
```

## Troubleshooting

### Database Connection Issues

```bash
# Check if PostgreSQL is running
docker ps

# Check logs
make db-logs

# Restart database
make db-restart
```

### No Data Showing

```bash
# Refresh materialized views
make refresh-views

# Check what data exists
psql $DATABASE_URL -c "SELECT COUNT(*) FROM events;"
```

### GraphQL Issues

```bash
# Regenerate GraphQL code
make graphql-deps
make graphql-generate
```

## License

MIT

## Contributing

Pull requests welcome! Please ensure:

- Code follows functional programming patterns
- Tests pass (`go test ./...`)
- GraphQL schema changes are documented