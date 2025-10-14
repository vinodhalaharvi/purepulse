# purepulse


# 🎯 FINAL Complete Type Classification Matrix + Data Flow

## 📊 **Complete Type Classification Matrix (Source → Presentation)**

### **LAYER 0: CONFIGURATION (Bootstrap)**

| Type | Category | Algebra | Purpose | Use purekernels? | Implementation | Purity |
|------|----------|---------|---------|------------------|----------------|--------|
| `AppConfig` | Config | None | Global settings (DB URL, etc.) | ❌ Custom | Concrete struct | ✅ Pure |
| `PlatformConfig` | Config | None | API credentials per platform | ❌ Custom | Concrete struct | ✅ Pure |
| `TimeRange` | Config | None | Query time boundaries | ❌ Custom | Concrete struct | ✅ Pure |
| `UserID` | Type Alias | None | User identifier | ❌ Custom | `type UserID string` | ✅ Pure |

---

### **LAYER 1: EXTERNAL SOURCES (API Connectors)**

| Type | Category | Algebra | Purpose | Use purekernels? | Implementation | Purity |
|------|----------|---------|---------|------------------|----------------|--------|
| `Platform` | Enum | None | `slack`, `github`, `jira`, `zoom` | ❌ Custom | `type Platform string` | ✅ Pure |
| `EventType` | Enum | None | `message`, `commit`, `issue`, `meeting` | ❌ Custom | `type EventType string` | ✅ Pure |
| `Connector` | Interface | None | Fetches from external API | ❌ Custom | Struct + function | ⚠️ Impure (I/O) |
| `ConnectorConfig` | Config | None | Rate limits, retry settings | ❌ Custom | Concrete struct | ✅ Pure |
| `RawAPIResponse` | Data | None | Platform-specific response | ❌ Custom | Transient type | ✅ Pure |
| `FetchError` | Error | None | Individual connector failure | ❌ Custom | Concrete struct | ✅ Pure |
| `[]FetchError` | Collection | Monoid (List) | Error accumulation | ✅ **purekernels** | `monoid.ListMonoid[FetchError]` | ✅ Pure |

---

### **LAYER 2: NORMALIZATION (Domain Events)**

| Type | Category | Algebra | Purpose | Use purekernels? | Implementation | Purity |
|------|----------|---------|---------|------------------|----------------|--------|
| `Event` | Data | None | Normalized cross-platform event | ❌ Custom | Concrete struct | ✅ Pure |
| `[]Event` | Collection | Monoid (List) | Event accumulation | ✅ **purekernels** | `monoid.ListMonoid[Event]` | ✅ Pure |
| `UserActivity` | Aggregate | Monoid | Per-platform event collections | ❌ Custom | Concrete struct | ✅ Pure |
| `UserActivityMonoid` | Monoid | Monoid | Combines partial activities | ❌ Custom | Implement `Monoid[UserActivity]` | ✅ Pure |
| `FetchMetadata` | Metrics | Monoid | API stats (latency, calls) | ❌ Custom | Concrete struct | ✅ Pure |
| `FetchMetadataMonoid` | Monoid | Monoid (Sum) | Combines metadata | ❌ Custom | Implement `Monoid[FetchMetadata]` | ✅ Pure |
| `FetchResult` | Result | Monoid | Events + Errors + Metadata | ❌ Custom | Concrete struct | ✅ Pure |
| `FetchResultMonoid` | Monoid | Monoid | Combines fetch results | ❌ Custom | Implement `Monoid[FetchResult]` | ✅ Pure |

---

### **LAYER 3: APPLICATIVE COMPOSITION (Validation, Planning, Parallel Fetch)**

| Type | Category | Algebra | Purpose | Use purekernels? | Implementation | Purity |
|------|----------|---------|---------|------------------|----------------|--------|
| `Validation[[]Error, T]` | Applicative | Validation | Input validation (accumulates errors) | ✅ **purekernels** | `either.Validation[[]Error, T]` | ✅ Pure |
| `ValidationError` | Error | None | Input validation failure | ❌ Custom | Concrete struct | ✅ Pure |
| `ExecutionPlan` | Plan | Monoid | Describes fetch strategy | ❌ Custom | Concrete struct | ✅ Pure |
| `PlanStep` | Plan | None | Individual plan step | ❌ Custom | Concrete struct | ✅ Pure |
| `CostEstimate` | Metrics | Monoid (Sum) | Estimated API calls, time | ❌ Custom | Concrete struct | ✅ Pure |
| `Const[ExecutionPlan, T]` | Applicative | Const | Build plan without executing | ✅ **purekernels** | `functor.Const[ExecutionPlan, T]` | ✅ Pure |
| `Concurrent[FetchResult]` | Applicative | Concurrent | Parallel fetch from 4 sources | ✅ **purekernels** | `functor.Concurrent[FetchResult]` | ✅ Pure (lazy) |
| `Writer[[]string, T]` | Applicative | Writer | Audit trail accumulation | ✅ **purekernels** | `effect.Writer[[]string, T]` | ✅ Pure (lazy) |

---

### **LAYER 4: DATABASE BOUNDARY (Reader + Writer)**

| Type | Category | Algebra | Purpose | Use purekernels? | Implementation | Purity |
|------|----------|---------|---------|------------------|----------------|--------|
| `Reader[*sql.DB, T]` | Applicative | Reader | DB connection injection | ✅ **purekernels** | `effect.Reader[*sql.DB, T]` | ✅ Pure (lazy) |
| `DBQuery[T]` | Type Alias | Reader | Query needing DB connection | ✅ **purekernels** | `type DBQuery[T] = Reader[*sql.DB, T]` | ✅ Pure (lazy) |
| `Writer[[]string, T]` | Applicative | Writer | Query audit logging | ✅ **purekernels** | `effect.Writer[[]string, T]` | ✅ Pure (lazy) |
| `AuditedDBQuery[T]` | Composed | Reader+Writer | Query with audit trail | ✅ **purekernels** | `Reader[*sql.DB, Writer[[]string, T]]` | ✅ Pure (lazy) |
| `DBTransaction[T]` | Type Alias | Reader | Transaction-scoped query | ✅ **purekernels** | `type DBTransaction[T] = Reader[*sql.Tx, T]` | ✅ Pure (lazy) |
| `InsertResult` | Result | None | Rows inserted + errors | ❌ Custom | Concrete struct | ✅ Pure |
| `QueryResult[T]` | Result | Either | Query output or error | ✅ **purekernels** | `either.Either[DBError, T]` | ✅ Pure |
| `DBError` | Error | None | Database operation failure | ❌ Custom | Concrete struct | ✅ Pure |

---

### **LAYER 5: STORAGE (Postgres - Impure I/O)**

| Type | Category | Algebra | Purpose | Use purekernels? | Implementation | Purity |
|------|----------|---------|---------|------------------|----------------|--------|
| `*sql.DB` | Resource | None | Database connection pool | ❌ stdlib | Standard library | ⚠️ Impure |
| `*sql.Tx` | Resource | None | Database transaction | ❌ stdlib | Standard library | ⚠️ Impure |
| `events` table | Schema | None | Normalized event storage | ❌ SQL DDL | Postgres table | ⚠️ Impure |
| `correlations` table | Schema | None | Detected patterns cache | ❌ SQL DDL | Postgres table | ⚠️ Impure |

---

### **LAYER 6: ANALYTICS (Query & Aggregate)**

| Type | Category | Algebra | Purpose | Use purekernels? | Implementation | Purity |
|------|----------|---------|---------|------------------|----------------|--------|
| `StructuredMetrics` | Aggregate | Monoid | Counts, scores, durations | ❌ Custom | Concrete struct | ✅ Pure |
| `StructuredMetricsMonoid` | Monoid | Monoid (Sum/Avg) | Combines metrics | ❌ Custom | Implement using purekernels monoids | ✅ Pure |
| `Sum[T]` | Monoid | Monoid (Sum) | Additive metrics | ✅ **purekernels** | `monoid.SumMonoid[T]` | ✅ Pure |
| `Avg` | Monoid | Monoid (Avg) | Weighted averages | ✅ **purekernels** | `monoid.AvgMonoid` | ✅ Pure |
| `Max[T]` | Monoid | Monoid (Max) | Maximum values | ✅ **purekernels** | `monoid.MaxMonoid[T]` | ✅ Pure |
| `Min[T]` | Monoid | Monoid (Min) | Minimum values | ✅ **purekernels** | `monoid.MinMonoid[T]` | ✅ Pure |

---

### **LAYER 7: CORRELATION (Pattern Detection with ZipList)**

| Type | Category | Algebra | Purpose | Use purekernels? | Implementation | Purity |
|------|----------|---------|---------|------------------|----------------|--------|
| `Correlation` | Analysis | None | Detected cross-platform pattern | ❌ Custom | Concrete struct | ✅ Pure |
| `CorrelationType` | Enum | None | Pattern type (slack→github, etc.) | ❌ Custom | `type CorrelationType string` | ✅ Pure |
| `[]Correlation` | Collection | Monoid (List) | All detected patterns | ✅ **purekernels** | `monoid.ListMonoid[Correlation]` | ✅ Pure |
| `ZipList[Event]` | Applicative | ZipList | Time-aligned event pairing | ✅ **purekernels** | `functor.ZipList[Event]` | ✅ Pure |
| `CorrelationWindow` | Config | None | Time window (e.g., 1 hour) | ❌ Custom | Concrete struct | ✅ Pure |
| `CorrelationScore` | Metrics | None | Confidence score (0.0-1.0) | ❌ Custom | `type CorrelationScore float64` | ✅ Pure |

---

### **LAYER 8: LLM SUMMARIZATION (AI Insights)**

| Type | Category | Algebra | Purpose | Use purekernels? | Implementation | Purity |
|------|----------|---------|---------|------------------|----------------|--------|
| `LLMClient` | Service | None | Claude API wrapper | ❌ Custom | Concrete struct | ⚠️ Impure (I/O) |
| `Prompt` | Data | None | Generated prompt for LLM | ❌ Custom | Concrete struct | ✅ Pure |
| `AISummary` | Output | Monoid | LLM-generated insights | ❌ Custom | Concrete struct | ✅ Pure |
| `AISummaryMonoid` | Monoid | Monoid (Concat) | Combines summaries | ❌ Custom | Implement using purekernels | ✅ Pure |
| `LLMMetadata` | Metrics | Monoid (Sum) | Tokens, latency, model | ❌ Custom | Concrete struct | ✅ Pure |
| `LLMError` | Error | None | LLM API failure | ❌ Custom | Concrete struct | ✅ Pure |
| `Concurrent[AISummary]` | Applicative | Concurrent | Parallel LLM calls | ✅ **purekernels** | `functor.Concurrent[AISummary]` | ✅ Pure (lazy) |

---

### **LAYER 9: FINAL AGGREGATION (Complete Summary)**

| Type | Category | Algebra | Purpose | Use purekernels? | Implementation | Purity |
|------|----------|---------|---------|------------------|----------------|--------|
| `UserSummary` | Aggregate | Monoid | Complete user analysis | ❌ Custom | Concrete struct | ✅ Pure |
| `UserSummaryMonoid` | Monoid | Monoid | Combines all layers | ❌ Custom | Implement using purekernels | ✅ Pure |
| `TeamSummary` | Aggregate | Monoid | Multi-user aggregation | ❌ Custom | Concrete struct | ✅ Pure |
| `TeamSummaryMonoid` | Monoid | Monoid | Combines user summaries | ❌ Custom | Implement using purekernels | ✅ Pure |
| `AuditLog` | Collection | Monoid (List) | Complete audit trail | ✅ **purekernels** | `monoid.ListMonoid[string]` | ✅ Pure |
| `Metadata` | Info | None | Timestamps, versions | ❌ Custom | Concrete struct | ✅ Pure |

---

### **LAYER 10: PRESENTATION (REST API - Applicatives)**

| Type | Category | Algebra | Purpose | Use purekernels? | Implementation | Purity |
|------|----------|---------|---------|------------------|----------------|--------|
| `HTTPRequest` | Input | None | HTTP request wrapper | ❌ Custom | Wraps `*http.Request` | ⚠️ Impure |
| `HTTPResponse` | Output | None | HTTP response wrapper | ❌ Custom | Concrete struct | ✅ Pure |
| `Reader[*http.Request, T]` | Applicative | Reader | Extract data from request | ✅ **purekernels** | `effect.Reader[*http.Request, T]` | ✅ Pure (lazy) |
| `Writer[HTTPResponse, T]` | Applicative | Writer | Build response + logs | ✅ **purekernels** | `effect.Writer[HTTPResponse, T]` | ✅ Pure (lazy) |
| `HTTPHandler[T]` | Composed | Reader+Writer | Complete HTTP handler | ✅ **purekernels** | `Reader[*http.Request, Writer[HTTPResponse, T]]` | ✅ Pure (lazy) |
| `RouteParams` | Data | None | URL parameters | ❌ Custom | Extracted from gorilla | ✅ Pure |
| `QueryParams` | Data | None | Query string parameters | ❌ Custom | Extracted from gorilla | ✅ Pure |
| `Validation[[]Error, T]` | Applicative | Validation | Validate request input | ✅ **purekernels** | `either.Validation[[]Error, T]` | ✅ Pure |
| `HTTPError` | Error | None | HTTP-level errors | ❌ Custom | Concrete struct | ✅ Pure |
| `StatusCode` | Enum | None | HTTP status codes | ❌ Custom | `type StatusCode int` | ✅ Pure |

---

### **LAYER 10.5: JSON ENCODING (Applicatives)**

| Type | Category | Algebra | Purpose | Use purekernels? | Implementation | Purity |
|------|----------|---------|---------|------------------|----------------|--------|
| `JSONEncoder[T]` | Applicative | Validation | Encode to JSON with validation | ✅ **purekernels** | `Validation[[]Error, json.RawMessage]` | ✅ Pure |
| `JSONDecoder[T]` | Applicative | Validation | Decode from JSON with validation | ✅ **purekernels** | `Validation[[]Error, T]` | ✅ Pure |
| `EncodingError` | Error | None | JSON encoding failure | ❌ Custom | Concrete struct | ✅ Pure |
| `DecodingError` | Error | None | JSON decoding failure | ❌ Custom | Concrete struct | ✅ Pure |

---

## 🔄 **Complete Data Flow (Source → Presentation)**

```
┌─────────────────────────────────────────────────────────────────────────┐
│ LAYER 0: CONFIGURATION (Bootstrap)                                     │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │
│ Load: AppConfig, PlatformConfig                                        │
│ Input: UserID, TimeRange (from HTTP request)                           │
└─────────────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────────────┐
│ LAYER 1: EXTERNAL SOURCES (Impure I/O - API Connectors)                │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │
│ Slack Connector  → RawAPIResponse → FetchError?                        │
│ GitHub Connector → RawAPIResponse → FetchError?                        │
│ Jira Connector   → RawAPIResponse → FetchError?                        │
│ Zoom Connector   → RawAPIResponse → FetchError?                        │
└─────────────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────────────┐
│ LAYER 2: NORMALIZATION (Pure)                                          │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │
│ RawAPIResponse → Event (normalized across platforms)                   │
│ []Event (per platform) → UserActivity                                  │
│ FetchErrors accumulated → []FetchError (ListMonoid)                    │
│ FetchMetadata (API stats) → Monoid combine                            │
│ Result: FetchResult = UserActivity + []FetchError + FetchMetadata     │
└─────────────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────────────┐
│ LAYER 3: APPLICATIVE COMPOSITION (Pure, Lazy)                          │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │
│ Validation[[]Error, Input] → Validate UserID, TimeRange               │
│ Const[ExecutionPlan, _]    → Build fetch plan (which sources, cost)   │
│ Concurrent[FetchResult]    → Parallel fetch from 4 connectors          │
│   ├─ SlackConnector   ┐                                                │
│   ├─ GitHubConnector  ├─→ Apply (Concurrent) → Combined FetchResult   │
│   ├─ JiraConnector    │                                                │
│   └─ ZoomConnector    ┘                                                │
│ Writer[[]string, _]        → Accumulate audit log                      │
│ Result: FetchResult with audit trail                                   │
└─────────────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────────────┐
│ LAYER 4: DATABASE BOUNDARY (Pure, Lazy - Reader + Writer)              │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │
│ Reader[*sql.DB, _] → Inject DB connection (pure until .Run(db))       │
│ Writer[[]string, _] → Accumulate DB operation logs                     │
│                                                                         │
│ INSERT Operations (Reader + Writer):                                   │
│   AuditedDBQuery[InsertResult] = Reader[*sql.DB, Writer[[]string, _]] │
│   flattenEvents(FetchResult) → []Event → INSERT INTO events           │
│   Logs: ["insert_started: 120 events", "insert_completed: 23ms"]     │
│                                                                         │
│ SELECT Operations (Reader + Writer):                                   │
│   AuditedDBQuery[UserActivity] = Reader[*sql.DB, Writer[[]string, _]] │
│   QueryUserEvents per platform → combine with Monoid                  │
│   Logs: ["query_started: user=alice", "query_completed: 50 events"]  │
└─────────────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────────────┐
│ LAYER 5: STORAGE (Impure I/O - Postgres)                               │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │
│ Execute: .Run(db) ← I/O boundary (all Reader/Writer execute here)     │
│                                                                         │
│ events table:                                                           │
│   ├─ INSERT: Normalized events from all platforms                     │
│   └─ SELECT: Query events by user_id, source, timestamp               │
│                                                                         │
│ Result: Persisted events + query results                              │
└─────────────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────────────┐
│ LAYER 6: ANALYTICS (Pure - SQL Aggregation + Monoids)                  │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │
│ []Event → StructuredMetrics (using SumMonoid, AvgMonoid)              │
│   ├─ Count messages/commits/issues (SumMonoid[int])                   │
│   ├─ Average code quality (AvgMonoid)                                 │
│   ├─ Max meeting duration (MaxMonoid)                                 │
│   └─ Combine with StructuredMetricsMonoid                             │
│                                                                         │
│ SQL GROUP BY queries:                                                  │
│   ├─ Per user  (user_id)                                              │
│   ├─ Per team  (team_id)                                              │
│   └─ Per source (source)                                              │
│                                                                         │
│ Result: StructuredMetrics aggregate                                    │
└─────────────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────────────┐
│ LAYER 7: CORRELATION (Pure - ZipList Applicative)                      │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │
│ Time-align events from different platforms:                            │
│   ZipList[Event] (Slack) ⊗ ZipList[Event] (GitHub)                    │
│     → Detect patterns within CorrelationWindow (1 hour)               │
│     → Generate Correlation with confidence score                       │
│                                                                         │
│ Detected Patterns:                                                     │
│   ├─ slack_to_github: Message → Commit (15 min avg)                  │
│   ├─ jira_to_github:  Issue → PR (2 hour avg)                        │
│   └─ zoom_to_action:  Meeting → Activity spike                       │
│                                                                         │
│ Result: []Correlation (ListMonoid[Correlation])                       │
└─────────────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────────────┐
│ LAYER 8: LLM SUMMARIZATION (Impure I/O - Concurrent Applicative)       │
│ ━━━━━━━━━━━━━━━━━━━━━━â━━ │
│ Build Prompts (Pure):                     etrics, []Correlation) → Prompt              │
│       pplicative):                          │
│   Concurrent[AInt[AISummary] (Collaboration) ┘                             │
│                                                                         │
│ Combine Results (AISummaryMonoid):                                    │
│   Overview:        Concat with separator                              │
│   Highlights:      Append lists                                        │
│   Recommendations: Append lists                                                                                            â
┌─────────────────────────────────────────────────────────────────────────┐
│ LAYER 9: FINAL AGGREGATION (Pure - UserSummaryMonoid)                  │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │
│ Combine all layers using UserSummaryMonoid:                           │
│   UserSummary {                                  "alice"                                             │
âys                                        │
│     Activi  Metrics:      StructuredMetrics (from Layer 6)              ummary (from Layer 8)                            │
│     AuditLog:     []string (accumulated from all layers)              │
│     GeneratedAt:  timestamp                                           │
│   }                                                                    │
│                                                                         │
│ Monoid properties allow combining partial summaries          m aggregation: fold over []UserSummary)                 │
─────────────────────â━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │
│ JSONEncoder[UserSummary] = Validation[[]Error, json.RawMessage]       │
│   ├─ Validate struct is serializable                                  │
│   ├─ Marshal to JSON                                                  │
│   └─ Return Valid(json) or Invalid([]EncodingError)                   │
│                                                                         │
│ Result: Validation[[]Error, json.RawMessage]                          │
└───────────────â│
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │
│ HTTPHandler[UserSummary] = Reader[*http.Request, Writer[Response, _]] │
│                                                                         │
│ Request Processing (Reader Applicative):                               │
│   Reader[*http.Request, RouteParams]  → Extract user_id from URL      │
│   Reader[*http.Request, QueryParams]  → Extract from, to dates        │
│   Re                                                           │
e):                                   │
│   Validation[[]Er[[]Error, TimeRange] → Validate date range               │
│                                                                         │
│ Business Logic (Pure):                                                 │
│   Call analytics pipeline → UserSummary                               │
│                                                                         │
│ Response Building (Writer Applicative):                                │
│   Writer[HTTPResponse, UserSummary]                                    │
│     Value: UserSummary                                                 │
│     Log:   ["request_received", "validation_passenerated", "response_sent"]                      │
                                 │
│ JSON Encoding (Validation):                                            │
│   JSONEncoder[UserSummary].Run() → json.RawMessage                    │
│                                                                         │
│ Execute at Edge (Impure I/O):                                          │
│   httpHandler.Run(request) → (HTTPResponse, []string)                 │
│   Write response to http.ResponseWriter                               │
└───────────────────────────────────â━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │
│ Route: GET /summary/user/{id}?from=...&to=...                         │
│                                                                         │
│ Wrapper (Thin adapter):                                                │
│   func(w http.ResponseWriter, r *http.Request) {                      │
│     response, l────────────────────────┘
```

---

## 🎯 **Applicative Composition Points**

| Layer | Applicatives Used | Purpose |
|-------|------------------|---------|
| **Layer 3: Composition** | `Validation`, `Const`, `Concurrent`, `Writer` | Validate input, plan execution, parallel fetch, audit |
| **Layer 4: DB Boundary** | `Reader`, `Writer` | Inject connection, accumulate logs |
| **Layer 7: Correlation** | `ZipList` | Time-align events from different sources |
| **Layer 8: LLM** | `Concurrent` | Parallel LLM API calls |
| **Layer 9: Aggregation** | Monoids (via `Foldable`) | Combine partial results |
| **Layer 10: JSON** | `Validation` | Encode/decode with error accumulation |
| **Layer 10: HTTP** | `Reader`, `Writer`, `Validation` | Extract request, build response, validate input |

---




# 🎯 Complete Domain Types (No Implementation)

Let me research the actual data structures from each platform to define accurate domain types.

---

## 📊 **LAYER 0: CONFIGURATION**

```go
package domain

import (
    "time"
)

// ============================================================================
// BOOTSTRAP CONFIGURATION
// ============================================================================

// AppConfig holds global application settings
type AppConfig struct {
    DatabaseURL      string
    ServerPort       int
    Environment      string // "development" | "staging" | "production"
    LogLevel         string
    EnableMetrics    bool
    EnableProfiling  bool
}

// PlatformConfig holds credentials for external APIs
type PlatformConfig struct {
    Slack  SlackConfig
    GitHub GitHubConfig
    Jira   JiraConfig
    Zoom   ZoomConfig
}

// SlackConfig for Slack API
type SlackConfig struct {
    Token            string
    AppToken         string // For socket mode
    SigningSecret    string
    WorkspaceID      string
    RateLimit        RateLimitConfig
    RetryConfig      RetryConfig
}

// GitHubConfig for GitHub API
type GitHubConfig struct {
    Token           string
    Organization    string
    RateLimit       RateLimitConfig
    RetryConfig     RetryConfig
    GraphQLEndpoint string
    RESTEndpoint    string
}

// JiraConfig for Jira API
type JiraConfig struct {
    URL          string // e.g., "https://company.atlassian.net"
    Email        string
    APIToken     string
    CloudID      string
    RateLimit    RateLimitConfig
    RetryConfig  RetryConfig
}

// ZoomConfig for Zoom API
type ZoomConfig struct {
    AccountID    string
    ClientID     string
    ClientSecret string
    RateLimit    RateLimitConfig
    RetryConfig  RetryConfig
}

// RateLimitConfig for API rate limiting
type RateLimitConfig struct {
    RequestsPerSecond int
    BurstSize         int
    MaxRetries        int
}

// RetryConfig for failed API calls
type RetryConfig struct {
    MaxRetries      int
    InitialDelay    time.Duration
    MaxDelay        time.Duration
    BackoffFactor   float64
}

// TimeRange for queries
type TimeRange struct {
    Start time.Time
    End   time.Time
}

// UserID identifies a user across platforms
type UserID string

// TeamID identifies a team
type TeamID string
```

---

## 📊 **LAYER 1: EXTERNAL SOURCES (Platform Enums & Connectors)**

```go
// ============================================================================
// PLATFORM IDENTIFIERS
// ============================================================================

// Platform enum for supported platforms
type Platform string

const (
    PlatformSlack  Platform = "slack"
    PlatformGitHub Platform = "github"
    PlatformJira   Platform = "jira"
    PlatformZoom   Platform = "zoom"
)

// EventType enum for normalized event types
type EventType string

const (
    // Slack event types
    EventSlackMessage       EventType = "slack_message"
    EventSlackReaction      EventType = "slack_reaction"
    EventSlackFileUpload    EventType = "slack_file_upload"
    EventSlackChannelJoin   EventType = "slack_channel_join"
    EventSlackThreadReply   EventType = "slack_thread_reply"
    
    // GitHub event types
    EventGitHubCommit       EventType = "github_commit"
    EventGitHubPR           EventType = "github_pull_request"
    EventGitHubPRReview     EventType = "github_pr_review"
    EventGitHubPRComment    EventType = "github_pr_comment"
    EventGitHubIssue        EventType = "github_issue"
    EventGitHubIssueComment EventType = "github_issue_comment"
    EventGitHubRelease      EventType = "github_release"
    
    // Jira event types
    EventJiraIssueCreated   EventType = "jira_issue_created"
    EventJiraIssueUpdated   EventType = "jira_issue_updated"
    EventJiraIssueClosed    EventType = "jira_issue_closed"
    EventJiraComment        EventType = "jira_comment"
    EventJiraTransition     EventType = "jira_transition"
    EventJiraSprint         EventType = "jira_sprint"
    
    // Zoom event types
    EventZoomMeeting        EventType = "zoom_meeting"
    EventZoomWebinar        EventType = "zoom_webinar"
    EventZoomRecording      EventType = "zoom_recording"
    EventZoomChat           EventType = "zoom_chat"
)

// ============================================================================
// SLACK RAW DATA STRUCTURES (Based on Slack API)
// ============================================================================

// SlackMessage from Slack API conversations.history
type SlackMessage struct {
    Type      string    `json:"type"`
    User      string    `json:"user"`
    Text      string    `json:"text"`
    Timestamp string    `json:"ts"` // Slack's unique timestamp format
    Channel   string    `json:"channel"`
    ThreadTS  string    `json:"thread_ts,omitempty"`
    ReplyCount int      `json:"reply_count,omitempty"`
    Reactions []SlackReaction `json:"reactions,omitempty"`
    Files     []SlackFile     `json:"files,omitempty"`
    Edited    *SlackEdited    `json:"edited,omitempty"`
    Blocks    []SlackBlock    `json:"blocks,omitempty"`
}

// SlackReaction represents emoji reactions
type SlackReaction struct {
    Name  string   `json:"name"`
    Count int      `json:"count"`
    Users []string `json:"users"`
}

// SlackFile represents uploaded files
type SlackFile struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Mimetype  string `json:"mimetype"`
    Filetype  string `json:"filetype"`
    Size      int    `json:"size"`
    URLPrivate string `json:"url_private"`
}

// SlackEdited represents message edit metadata
type SlackEdited struct {
    User      string `json:"user"`
    Timestamp string `json:"ts"`
}

// SlackBlock represents Slack Block Kit elements
type SlackBlock struct {
    Type     string                 `json:"type"`
    BlockID  string                 `json:"block_id,omitempty"`
    Elements []interface{}          `json:"elements,omitempty"`
    Fields   []interface{}          `json:"fields,omitempty"`
}

// SlackUser from Slack API users.info
type SlackUser struct {
    ID       string           `json:"id"`
    TeamID   string           `json:"team_id"`
    Name     string           `json:"name"`
    RealName string           `json:"real_name"`
    Email    string           `json:"profile.email"`
    IsBot    bool             `json:"is_bot"`
    IsAdmin  bool             `json:"is_admin"`
    TZ       string           `json:"tz"`
    Profile  SlackUserProfile `json:"profile"`
}

// SlackUserProfile contains user profile details
type SlackUserProfile struct {
    DisplayName string `json:"display_name"`
    Email       string `json:"email"`
    Phone       string `json:"phone"`
    Title       string `json:"title"`
    StatusText  string `json:"status_text"`
    StatusEmoji string `json:"status_emoji"`
}

// SlackChannel from Slack API conversations.list
type SlackChannel struct {
    ID             string   `json:"id"`
    Name           string   `json:"name"`
    IsChannel      bool     `json:"is_channel"`
    IsGroup        bool     `json:"is_group"`
    IsIM           bool     `json:"is_im"`
    IsPrivate      bool     `json:"is_private"`
    Created        int64    `json:"created"`
    NumMembers     int      `json:"num_members"`
    Topic          SlackTopic `json:"topic"`
    Purpose        SlackPurpose `json:"purpose"`
}

// SlackTopic for channel topic
type SlackTopic struct {
    Value   string `json:"value"`
    Creator string `json:"creator"`
    LastSet int64  `json:"last_set"`
}

// SlackPurpose for channel purpose
type SlackPurpose struct {
    Value   string `json:"value"`
    Creator string `json:"creator"`
    LastSet int64  `json:"last_set"`
}

// ============================================================================
// GITHUB RAW DATA STRUCTURES (Based on GitHub API v3 & GraphQL)
// ============================================================================

// GitHubCommit from GitHub API repos/{owner}/{repo}/commits
type GitHubCommit struct {
    SHA       string              `json:"sha"`
    NodeID    string              `json:"node_id"`
    Commit    GitHubCommitDetail  `json:"commit"`
    Author    *GitHubUser         `json:"author"`
    Committer *GitHubUser         `json:"committer"`
    Parents   []GitHubCommitRef   `json:"parents"`
    Stats     *GitHubCommitStats  `json:"stats"`
    Files     []GitHubCommitFile  `json:"files"`
    URL       string              `json:"url"`
    HTMLURL   string              `json:"html_url"`
}

// GitHubCommitDetail contains commit message and metadata
type GitHubCommitDetail struct {
    Author       GitHubCommitAuthor `json:"author"`
    Committer    GitHubCommitAuthor `json:"committer"`
    Message      string             `json:"message"`
    Tree         GitHubTree         `json:"tree"`
    CommentCount int                `json:"comment_count"`
    Verification GitHubVerification `json:"verification"`
}

// GitHubCommitAuthor contains author info
type GitHubCommitAuthor struct {
    Name  string    `json:"name"`
    Email string    `json:"email"`
    Date  time.Time `json:"date"`
}

// GitHubCommitStats contains line change statistics
type GitHubCommitStats struct {
    Additions int `json:"additions"`
    Deletions int `json:"deletions"`
    Total     int `json:"total"`
}

// GitHubCommitFile represents files changed in commit
type GitHubCommitFile struct {
    Filename  string `json:"filename"`
    Status    string `json:"status"` // "added", "removed", "modified", "renamed"
    Additions int    `json:"additions"`
    Deletions int    `json:"deletions"`
    Changes   int    `json:"changes"`
    Patch     string `json:"patch"`
}

// GitHubCommitRef for parent commits
type GitHubCommitRef struct {
    SHA string `json:"sha"`
    URL string `json:"url"`
}

// GitHubTree represents git tree
type GitHubTree struct {
    SHA string `json:"sha"`
    URL string `json:"url"`
}

// GitHubVerification for commit signature
type GitHubVerification struct {
    Verified  bool   `json:"verified"`
    Reason    string `json:"reason"`
    Signature string `json:"signature"`
    Payload   string `json:"payload"`
}

// GitHubPullRequest from GitHub API pulls/{number}
type GitHubPullRequest struct {
    Number      int                  `json:"number"`
    State       string               `json:"state"` // "open", "closed"
    Title       string               `json:"title"`
    Body        string               `json:"body"`
    User        GitHubUser           `json:"user"`
    CreatedAt   time.Time            `json:"created_at"`
    UpdatedAt   time.Time            `json:"updated_at"`
    ClosedAt    *time.Time           `json:"closed_at"`
    MergedAt    *time.Time           `json:"merged_at"`
    MergeCommitSHA string            `json:"merge_commit_sha"`
    Assignees   []GitHubUser         `json:"assignees"`
    Reviewers   []GitHubUser         `json:"requested_reviewers"`
    Labels      []GitHubLabel        `json:"labels"`
    Head        GitHubPRBranch       `json:"head"`
    Base        GitHubPRBranch       `json:"base"`
    Draft       bool                 `json:"draft"`
    Merged      bool                 `json:"merged"`
    Mergeable   *bool                `json:"mergeable"`
    Comments    int                  `json:"comments"`
    Commits     int                  `json:"commits"`
    Additions   int                  `json:"additions"`
    Deletions   int                  `json:"deletions"`
    ChangedFiles int                 `json:"changed_files"`
}

// GitHubPRBranch represents PR head/base branch
type GitHubPRBranch struct {
    Label string          `json:"label"`
    Ref   string          `json:"ref"`
    SHA   string          `json:"sha"`
    User  GitHubUser      `json:"user"`
    Repo  GitHubRepository `json:"repo"`
}

// GitHubPRReview from GitHub API pulls/{number}/reviews
type GitHubPRReview struct {
    ID          int64      `json:"id"`
    User        GitHubUser `json:"user"`
    Body        string     `json:"body"`
    State       string     `json:"state"` // "APPROVED", "CHANGES_REQUESTED", "COMMENTED"
    SubmittedAt time.Time  `json:"submitted_at"`
    CommitID    string     `json:"commit_id"`
}

// GitHubIssue from GitHub API issues/{number}
type GitHubIssue struct {
    Number      int              `json:"number"`
    State       string           `json:"state"` // "open", "closed"
    Title       string           `json:"title"`
    Body        string           `json:"body"`
    User        GitHubUser       `json:"user"`
    Labels      []GitHubLabel    `json:"labels"`
    Assignees   []GitHubUser     `json:"assignees"`
    Milestone   *GitHubMilestone `json:"milestone"`
    Comments    int              `json:"comments"`
    CreatedAt   time.Time        `json:"created_at"`
    UpdatedAt   time.Time        `json:"updated_at"`
    ClosedAt    *time.Time       `json:"closed_at"`
    PullRequest *GitHubPRRef     `json:"pull_request,omitempty"`
}

// GitHubUser represents a GitHub user
type GitHubUser struct {
    Login     string `json:"login"`
    ID        int64  `json:"id"`
    NodeID    string `json:"node_id"`
    AvatarURL string `json:"avatar_url"`
    Type      string `json:"type"` // "User", "Bot", "Organization"
    SiteAdmin bool   `json:"site_admin"`
}

// GitHubLabel for issue/PR labels
type GitHubLabel struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Color string `json:"color"`
}

// GitHubMilestone for issue milestones
type GitHubMilestone struct {
    Number   int    `json:"number"`
    Title    string `json:"title"`
    State    string `json:"state"`
    DueOn    *time.Time `json:"due_on"`
}

// GitHubPRRef indicates if issue is a PR
type GitHubPRRef struct {
    URL string `json:"url"`
}

// GitHubRepository represents a repository
type GitHubRepository struct {
    ID       int64  `json:"id"`
    Name     string `json:"name"`
    FullName string `json:"full_name"`
    Private  bool   `json:"private"`
    Owner    GitHubUser `json:"owner"`
}

// ============================================================================
// JIRA RAW DATA STRUCTURES (Based on Jira Cloud REST API v3)
// ============================================================================

// JiraIssue from Jira API /rest/api/3/issue/{issueIdOrKey}
type JiraIssue struct {
    ID     string         `json:"id"`
    Key    string         `json:"key"`
    Self   string         `json:"self"`
    Fields JiraIssueFields `json:"fields"`
}

// JiraIssueFields contains all issue fields
type JiraIssueFields struct {
    Summary     string            `json:"summary"`
    Description *JiraDescription  `json:"description"`
    IssueType   JiraIssueType     `json:"issuetype"`
    Project     JiraProject       `json:"project"`
    Status      JiraStatus        `json:"status"`
    Priority    *JiraPriority     `json:"priority"`
    Resolution  *JiraResolution   `json:"resolution"`
    Assignee    *JiraUser         `json:"assignee"`
    Reporter    JiraUser          `json:"reporter"`
    Creator     JiraUser          `json:"creator"`
    Created     time.Time         `json:"created"`
    Updated     time.Time         `json:"updated"`
    ResolutionDate *time.Time     `json:"resolutiondate"`
    DueDate     *string           `json:"duedate"` // YYYY-MM-DD format
    Labels      []string          `json:"labels"`
    Components  []JiraComponent   `json:"components"`
    FixVersions []JiraVersion     `json:"fixVersions"`
    Sprint      *JiraSprint       `json:"sprint"`
    StoryPoints *float64          `json:"customfield_10016"` // Story points (custom field)
    TimeTracking *JiraTimeTracking `json:"timetracking"`
    Comment     *JiraCommentList  `json:"comment"`
    Worklog     *JiraWorklogList  `json:"worklog"`
    Parent      *JiraIssueRef     `json:"parent"`
    Subtasks    []JiraIssueRef    `json:"subtasks"`
}

// JiraDescription in ADF (Atlassian Document Format)
type JiraDescription struct {
    Type    string                 `json:"type"`
    Version int                    `json:"version"`
    Content []JiraDescriptionNode  `json:"content"`
}

// JiraDescriptionNode represents ADF nodes
type JiraDescriptionNode struct {
    Type    string                   `json:"type"`
    Content []JiraDescriptionNode    `json:"content,omitempty"`
    Text    string                   `json:"text,omitempty"`
    Attrs   map[string]interface{}   `json:"attrs,omitempty"`
}

// JiraIssueType represents issue type
type JiraIssueType struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    IconURL     string `json:"iconUrl"`
    Subtask     bool   `json:"subtask"`
}

// JiraProject represents a project
type JiraProject struct {
    ID   string `json:"id"`
    Key  string `json:"key"`
    Name string `json:"name"`
}

// JiraStatus represents issue status
type JiraStatus struct {
    ID             string              `json:"id"`
    Name           string              `json:"name"`
    StatusCategory JiraStatusCategory  `json:"statusCategory"`
}

// JiraStatusCategory for status grouping
type JiraStatusCategory struct {
    ID   int    `json:"id"`
    Key  string `json:"key"` // "new", "indeterminate", "done"
    Name string `json:"name"`
}

// JiraPriority represents issue priority
type JiraPriority struct {
    ID      string `json:"id"`
    Name    string `json:"name"`
    IconURL string `json:"iconUrl"`
}

// JiraResolution represents issue resolution
type JiraResolution struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
}

// JiraUser represents a Jira user
type JiraUser struct {
    AccountID    string `json:"accountId"`
    DisplayName  string `json:"displayName"`
    EmailAddress string `json:"emailAddress"`
    Active       bool   `json:"active"`
    TimeZone     string `json:"timeZone"`
    AccountType  string `json:"accountType"` // "atlassian", "app"
}

// JiraComponent represents project component
type JiraComponent struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

// JiraVersion represents fix version
type JiraVersion struct {
    ID          string     `json:"id"`
    Name        string     `json:"name"`
    Released    bool       `json:"released"`
    ReleaseDate *string    `json:"releaseDate"`
}

// JiraSprint represents agile sprint
type JiraSprint struct {
    ID            int        `json:"id"`
    Name          string     `json:"name"`
    State         string     `json:"state"` // "future", "active", "closed"
    StartDate     *time.Time `json:"startDate"`
    EndDate       *time.Time `json:"endDate"`
    CompleteDate  *time.Time `json:"completeDate"`
    OriginBoardID int        `json:"originBoardId"`
}

// JiraTimeTracking for time estimates
type JiraTimeTracking struct {
    OriginalEstimate         string `json:"originalEstimate"`
    RemainingEstimate        string `json:"remainingEstimate"`
    TimeSpent                string `json:"timeSpent"`
    OriginalEstimateSeconds  int    `json:"originalEstimateSeconds"`
    RemainingEstimateSeconds int    `json:"remainingEstimateSeconds"`
    TimeSpentSeconds         int    `json:"timeSpentSeconds"`
}

// JiraCommentList contains comments
type JiraCommentList struct {
    Comments []JiraComment `json:"comments"`
    Total    int           `json:"total"`
}

// JiraComment represents a comment
type JiraComment struct {
    ID      string           `json:"id"`
    Author  JiraUser         `json:"author"`
    Body    *JiraDescription `json:"body"`
    Created time.Time        `json:"created"`
    Updated time.Time        `json:"updated"`
}

// JiraWorklogList contains work logs
type JiraWorklogList struct {
    Worklogs []JiraWorklog `json:"worklogs"`
    Total    int           `json:"total"`
}

// JiraWorklog represents logged work
type JiraWorklog struct {
    ID               string    `json:"id"`
    Author           JiraUser  `json:"author"`
    Comment          string    `json:"comment"`
    Started          time.Time `json:"started"`
    TimeSpent        string    `json:"timeSpent"`
    TimeSpentSeconds int       `json:"timeSpentSeconds"`
}

// JiraIssueRef for parent/subtask references
type JiraIssueRef struct {
    ID     string              `json:"id"`
    Key    string              `json:"key"`
    Fields JiraIssueRefFields  `json:"fields"`
}

// JiraIssueRefFields for minimal issue info
type JiraIssueRefFields struct {
    Summary   string        `json:"summary"`
    Status    JiraStatus    `json:"status"`
    IssueType JiraIssueType `json:"issuetype"`
}

// JiraTransition represents status transition
type JiraTransition struct {
    ID         string     `json:"id"`
    Name       string     `json:"name"`
    To         JiraStatus `json:"to"`
    HasScreen  bool       `json:"hasScreen"`
    IsGlobal   bool       `json:"isGlobal"`
    IsInitial  bool       `json:"isInitial"`
    IsConditional bool    `json:"isConditional"`
}

// ============================================================================
// ZOOM RAW DATA STRUCTURES (Based on Zoom API v2)
// ============================================================================

// ZoomMeeting from Zoom API /meetings/{meetingId}
type ZoomMeeting struct {
    UUID            string            `json:"uuid"`
    ID              int64             `json:"id"`
    HostID          string            `json:"host_id"`
    HostEmail       string            `json:"host_email"`
    Topic           string            `json:"topic"`
    Type            int               `json:"type"` // 1=instant, 2=scheduled, 3=recurring, 8=recurring fixed
    Status          string            `json:"status"` // "waiting", "started", "finished"
    StartTime       time.Time         `json:"start_time"`
    Duration        int               `json:"duration"` // minutes
    Timezone        string            `json:"timezone"`
    CreatedAt       time.Time         `json:"created_at"`
    JoinURL         string            `json:"join_url"`
    Agenda          string            `json:"agenda"`
    StartURL        string            `json:"start_url"`
    Password        string            `json:"password"`
    Settings        ZoomMeetingSettings `json:"settings"`
    Recurrence      *ZoomRecurrence   `json:"recurrence"`
}

// ZoomMeetingSettings for meeting configuration
type ZoomMeetingSettings struct {
    HostVideo             bool   `json:"host_video"`
    ParticipantVideo      bool   `json:"participant_video"`
    JoinBeforeHost        bool   `json:"join_before_host"`
    MuteUponEntry         bool   `json:"mute_upon_entry"`
    Watermark             bool   `json:"watermark"`
    AudioType             string `json:"audio"` // "both", "telephony", "voip"
    AutoRecording         string `json:"auto_recording"` // "local", "cloud", "none"
    WaitingRoom           bool   `json:"waiting_room"`
    BreakoutRoom          bool   `json:"breakout_room"`
    MeetingAuthentication bool   `json:"meeting_authentication"`
}

// ZoomRecurrence for recurring meetings
type ZoomRecurrence struct {
    Type           int    `json:"type"` // 1=Daily, 2=Weekly, 3=Monthly
    RepeatInterval int    `json:"repeat_interval"`
    WeeklyDays     string `json:"weekly_days"` // "1,2,3,4,5"
    MonthlyDay     int    `json:"monthly_day"`
    MonthlyWeek    int    `json:"monthly_week"`
    MonthlyWeekDay int    `json:"monthly_week_day"`
    EndTimes       int    `json:"end_times"`
    EndDateTime    *time.Time `json:"end_date_time"`
}

// ZoomParticipant from Zoom API /metrics/meetings/{meetingId}/participants
type ZoomParticipant struct {
    ID                string    `json:"id"`
    UserID            string    `json:"user_id"`
    Name              string    `json:"name"`
    UserEmail         string    `json:"user_email"`
    JoinTime          time.Time `json:"join_time"`
    LeaveTime         time.Time `json:"leave_time"`
    Duration          int       `json:"duration"` // seconds
    Attentiveness     int       `json:"attentiveness_score"` // 0-100
    FailoverInfo      string    `json:"failover"`
    Status            string    `json:"status"` // "in_meeting", "in_waiting_room"
    CameraEnabled     bool      `json:"camera"`
    MicrophoneEnabled bool      `json:"microphone"`
    ScreenShare       bool      `json:"share_screen"`
    Recording         bool      `json:"recording"`
    Role              string    `json:"role"` // "host", "co-host", "participant"
}

// ZoomWebinar from Zoom API /webinars/{webinarId}
type ZoomWebinar struct {
    UUID       string              `json:"uuid"`
    ID         int64               `json:"id"`
    HostID     string              `json:"host_id"`
    Topic      string              `json:"topic"`
    Type       int                 `json:"type"`
    StartTime  time.Time           `json:"start_time"`
    Duration   int                 `json:"duration"`
    Timezone   string              `json:"timezone"`
    CreatedAt  time.Time           `json:"created_at"`
    JoinURL    string              `json:"join_url"`
    Agenda     string              `json:"agenda"`
    Settings   ZoomWebinarSettings `json:"settings"`
}

// ZoomWebinarSettings for webinar configuration
type ZoomWebinarSettings struct {
    HostVideo            bool   `json:"host_video"`
    PanelistsVideo       bool   `json:"panelists_video"`
    PracticeSession      bool   `json:"practice_session"`
    HDVideo              bool   `json:"hd_video"`
    ApprovalType         int    `json:"approval_type"` // 0=auto, 1=manual, 2=no registration
    RegistrationType     int    `json:"registration_type"`
    Audio                string `json:"audio"`
    AutoRecording        string `json:"auto_recording"`
    OnDemand             bool   `json:"on_demand"`
    ShowShareButton      bool   `json:"show_share_button"`
    AllowMultipleDevices bool   `json:"allow_multiple_devices"`
}

// ZoomRecording from Zoom API /meetings/{meetingId}/recordings
type ZoomRecording struct {
    UUID           string                  `json:"uuid"`
    ID             int64                   `json:"id"`
    AccountID      string                  `json:"account_id"`
    HostID         string                  `json:"host_id"`
    Topic          string                  `json:"topic"`
    StartTime      time.Time               `json:"start_time"`
    Duration       int                     `json:"duration"`
    TotalSize      int64                   `json:"total_size"`
    RecordingCount int                     `json:"recording_count"`
    RecordingFiles []ZoomRecordingFile     `json:"recording_files"`
}

// ZoomRecordingFile represents individual recording file
type ZoomRecordingFile struct {
    ID             string    `json:"id"`
    MeetingID      string    `json:"meeting_id"`
    RecordingStart time.Time `json:"recording_start"`
    RecordingEnd   time.Time `json:"recording_end"`
    FileType       string    `json:"file_type"` // "MP4", "M4A", "TIMELINE", "TRANSCRIPT", "CHAT"
    FileSize       int64     `json:"file_size"`
    PlayURL        string    `json:"play_url"`
    DownloadURL    string    `json:"download_url"`
    Status         string    `json:"status"` // "completed", "processing"
    RecordingType  string    `json:"recording_type"` // "shared_screen_with_speaker_view", "audio_only", etc.
}

// ZoomChatMessage from Zoom API /chat/users/{userId}/messages
type ZoomChatMessage struct {
    ID        string    `json:"id"`
    Message   string    `json:"message"`
    Sender    string    `json:"sender"`
    ToContact string    `json:"to_contact"`
    ToChannel string    `json:"to_channel"`
    DateTime  time.Time `json:"date_time"`
    Timestamp int64     `json:"timestamp"`
}

// ============================================================================
// CONNECTOR INTERFACES & RESULTS
// ============================================================================

// Connector interface for platform fetchers
type Connector struct {
    Platform Platform
    Fetch    func(ctx context.Context, userID UserID, tr TimeRange) ([]Event, error)
}

// ConnectorConfig for connector-specific settings
type ConnectorConfig struct {
    Platform     Platform
    Enabled      bool
    RateLimit    RateLimitConfig
    RetryConfig  RetryConfig
    CacheConfig  CacheConfig
}

// CacheConfig for connector caching
type CacheConfig struct {
    Enabled bool
    TTL     time.Duration
    MaxSize int
}

// FetchError represents connector failure
type FetchError struct {
    Source       Platform
    Error        error
    Timestamp    time.Time
    Retryable    bool
    StatusCode   int    // HTTP status code if applicable
    RateLimited  bool
    RetryAfter   *time.Duration
}
```

---

## 📊 **LAYER 2: NORMALIZATION (Domain Events)**

```go
// ============================================================================
// NORMALIZED DOMAIN EVENTS
// ============================================================================

// Event is the normalized unit across all platforms
type Event struct {
    ID        string          `json:"id"`
    UserID    UserID          `json:"user_id"`
    Source    Platform        `json:"source"`
    Type      EventType       `json:"type"`
    Timestamp time.Time       `json:"timestamp"`
    Payload   json.RawMessage `json:"payload"` // Original platform data
    Metadata  EventMetadata   `json:"metadata"`
}

// EventMetadata contains derived metadata
type EventMetadata struct {
    // Common across all platforms
    Author       string            `json:"author,omitempty"`
    Channel      string            `json:"channel,omitempty"` // Slack channel, GitHub repo, Jira project
    ThreadID     string            `json:"thread_id,omitempty"`
    ParentID     string            `json:"parent_id,omitempty"`
    Tags         []string          `json:"tags,omitempty"`
    Participants []string          `json:"participants,omitempty"`
    
    // Metrics
    Size         int               `json:"size,omitempty"` // Lines of code, message length, etc.
    Duration     *time.Duration    `json:"duration,omitempty"` // Meeting duration, PR open time
    
    // Relationships
    RelatedEventIDs []string `json:"related_event_ids,omitempty"` // Related events, linked issues    
	
    // Platform-specific (kept for correlation)
    Extra        map[string]interface{} `json:"extra,omitempty"`
}

// UserActivity combines events from all platforms
type UserActivity struct {
    UserID    UserID        `json:"user_id"`
    TimeRange TimeRange     `json:"time_range"`
    
    // Per-platform events
    Slack     []Event       `json:"slack"`
    GitHub    []Event       `json:"github"`
    Jira      []Event       `json:"jira"`
    Zoom      []Event       `json:"zoom"`
    
    // Metadata
    Metadata  FetchMetadata `json:"metadata"`
}

// FetchMetadata tracks fetch performance
type FetchMetadata struct {
    TotalAPICalls  int                       `json:"total_api_calls"`
    TotalEvents    int                       `json:"total_events"`
    TotalLatencyMS int                       `json:"total_latency_ms"`
    CacheHits      int                       `json:"cache_hits"`
    FetchedAt      time.Time                 `json:"fetched_at"`
    SourceMetrics  map[Platform]SourceMetrics `json:"source_metrics"`
}

// SourceMetrics for per-platform stats
type SourceMetrics struct {
    APICalls     int           `json:"api_calls"`
    Events       int           `json:"events"`
    LatencyMS    int           `json:"latency_ms"`
    Success      bool          `json:"success"`
    Cached       bool          `json:"cached"`
    RateLimited  bool          `json:"rate_limited"`
    Error        string        `json:"error,omitempty"`
}

// FetchResult combines activity, errors, and metadata
type FetchResult struct {
    Activity UserActivity  `json:"activity"`
    Errors   []FetchError  `json:"errors"`
    Metadata FetchMetadata `json:"metadata"`
}
```

---

## 📊 **LAYER 3: VALIDATION & PLANNING**

```go
// ============================================================================
// VALIDATION & EXECUTION PLANNING
// ============================================================================

// ValidationError for input validation failures
type ValidationError struct {
    Field   string `json:"field"`
    Value   string `json:"value"`
    Message string `json:"message"`
    Code    string `json:"code"` // "required", "invalid_format", "out_of_range"
}

// ExecutionPlan describes fetch strategy
type ExecutionPlan struct {
    UserID       UserID          `json:"user_id"`
    TimeRange    TimeRange       `json:"time_range"`
    Sources      []Platform      `json:"sources"`
    Steps        []PlanStep      `json:"steps"`
    Estimated    CostEstimate    `json:"estimated"`
    CreatedAt    time.Time       `json:"created_at"`
}

// PlanStep represents individual fetch step
type PlanStep struct {
    Source       Platform        `json:"source"`
    Operation    string          `json:"operation"` // "fetch_messages", "fetch_commits", etc.
    EstimatedMS  int             `json:"estimated_ms"`
    EstimatedAPI int             `json:"estimated_api_calls"`
    CacheCheck   bool            `json:"cache_check"`
}

// CostEstimate for resource planning
type CostEstimate struct {
    TotalAPICalls int           `json:"total_api_calls"`
    TotalTimeMS   int           `json:"total_time_ms"`
    CacheHitRate  float64       `json:"cache_hit_rate"`
    RateLimitRisk bool          `json:"rate_limit_risk"`
}
```

---

## 📊 **LAYER 4: DATABASE TYPES**

```go
// ============================================================================
// DATABASE OPERATIONS
// ============================================================================

// InsertResult for bulk insert operations
type InsertResult struct {
    RowsInserted  int           `json:"rows_inserted"`
    RowsFailed    int           `json:"rows_failed"`
    Duration      time.Duration `json:"duration"`
    Errors        []DBError     `json:"errors,omitempty"`
}

// DBError for database operation failures
type DBError struct {
    Operation  string    `json:"operation"` // "insert", "select", "update"
    Table      string    `json:"table"`
    Query      string    `json:"query"`
    Error      error     `json:"error"`
    Timestamp  time.Time `json:"timestamp"`
    Retryable  bool      `json:"retryable"`
}
```

---

## 📊 **LAYER 6: ANALYTICS METRICS**

```go
// ============================================================================
// STRUCTURED ANALYTICS METRICS
// ============================================================================

// StructuredMetrics are computed from raw events
type StructuredMetrics struct {
    UserID    UserID    `json:"user_id"`
    TimeRange TimeRange `json:"time_range"`
    
    // Slack metrics
    SlackMessages      int     `json:"slack_messages"`
    SlackChannels      int     `json:"slack_channels"`
    SlackReactions     int     `json:"slack_reactions"`
    SlackFilesShared   int     `json:"slack_files_shared"`
    SlackThreadReplies int     `json:"slack_thread_replies"`
    
    // GitHub metrics
    GitHubCommits      int     `json:"github_commits"`
    GitHubPRs          int     `json:"github_prs"`
    GitHubPRsReviewed  int     `json:"github_prs_reviewed"`
    GitHubIssues       int     `json:"github_issues"`
    GitHubComments     int     `json:"github_comments"`
    GitHubLinesAdded   int     `json:"github_lines_added"`
    GitHubLinesDeleted int     `json:"github_lines_deleted"`
    CodeQuality        float64 `json:"code_quality"` // Derived score
    
    // Jira metrics
    JiraIssuesCreated  int     `json:"jira_issues_created"`
    JiraIssuesCompleted int    `json:"jira_issues_completed"`
    JiraStoryPoints    float64 `json:"jira_story_points"`
    JiraVelocity       float64 `json:"jira_velocity"` // Points per sprint
    JiraComments       int     `json:"jira_comments"`
    JiraTransitions    int     `json:"jira_transitions"`
    
    // Zoom metrics
    ZoomMeetings       int     `json:"zoom_meetings"`
    ZoomMinutes        int     `json:"zoom_minutes"`
    ZoomWebinars       int     `json:"zoom_webinars"`
    ZoomRecordings     int     `json:"zoom_recordings"`
    
    // Derived cross-platform metrics
    ActiveHours        float64 `json:"active_hours"`
    PeakHour           int     `json:"peak_hour"` // 0-23
    CollaborationScore float64 `json:"collaboration_score"` // 0-100
    
    // Metadata
    DataPoints         int     `json:"data_points"` // Total events processed
}
```

---

## 📊 **LAYER 7: CORRELATION**

```go
// ============================================================================
// PATTERN CORRELATION
// ============================================================================

// CorrelationType enum for detected patterns
type CorrelationType string

const (
    CorrelationSlackToGitHub  CorrelationType = "slack_to_github"
    CorrelationJiraToGitHub   CorrelationType = "jira_to_github"
    CorrelationZoomToActivity CorrelationType = "zoom_to_activity"
    CorrelationGitHubToJira   CorrelationType = "github_to_jira"
    CorrelationSlackToJira    CorrelationType = "slack_to_jira"
)

// Correlation represents detected cross-platform pattern
type Correlation struct {
    ID          string           `json:"id"`
    UserID      UserID           `json:"user_id"`
    Type        CorrelationType  `json:"type"`
    Events      []Event          `json:"events"` // Related events (source + target)
    Confidence  float64          `json:"confidence"` // 0.0 - 1.0
    Description string           `json:"description"`
    TimeDelta   time.Duration    `json:"time_delta"` // Time between correlated events
    Frequency   int              `json:"frequency"` // How many times pattern occurred
    DetectedAt  time.Time        `json:"detected_at"`
}

// CorrelationWindow for time-window correlation
type CorrelationWindow struct {
    Duration time.Duration `json:"duration"` // e.g., 1 hour
    Offset   time.Duration `json:"offset"`   // Slide window by this amount
}

// CorrelationScore for pattern strength
type CorrelationScore float64
```

---

## 📊 **LAYER 8: LLM SUMMARIZATION**

```go
// ============================================================================
// AI SUMMARIZATION
// ============================================================================

// Prompt for LLM input
type Prompt struct {
    Audience string `json:"audience"` // "executive" | "technical" | "team"
    Context  string `json:"context"`  // Generated from metrics + correlations
    MaxTokens int   `json:"max_tokens"`
}

// AISummary from LLM
type AISummary struct {
    Overview        string          `json:"overview"`
    Highlights      []string        `json:"highlights"`
    Recommendations []string        `json:"recommendations"`
    Metadata        LLMMetadata     `json:"metadata"`
}

// LLMMetadata for LLM call stats
type LLMMetadata struct {
    Model      string        `json:"model"`
    Tokens     int           `json:"tokens"`
    LatencyMS  int           `json:"latency_ms"`
    CachedPrompt bool        `json:"cached_prompt"`
    Cost       float64       `json:"cost"` // Estimated API cost
}

// LLMError for LLM API failures
type LLMError struct {
    Provider  string    `json:"provider"` // "anthropic", "openai"
    Error     error     `json:"error"`
    Timestamp time.Time `json:"timestamp"`
    Retryable bool      `json:"retryable"`
    StatusCode int      `json:"status_code"`
}
```

---

## 📊 **LAYER 9: FINAL AGGREGATION**

```go
// ============================================================================
// COMPLETE USER SUMMARY
// ============================================================================

// UserSummary is the final output
type UserSummary struct {
    UserID       UserID              `json:"user_id"`
    TimeRange    TimeRange           `json:"time_range"`
    
    // All layers combined
    Activity     UserActivity        `json:"activity"`
    Metrics      StructuredMetrics   `json:"metrics"`
    Correlations []Correlation       `json:"correlations"`
    AISummary    AISummary           `json:"ai_summary"`
    
    // Metadata
    GeneratedAt  time.Time           `json:"generated_at"`
    AuditLog     []string            `json:"audit_log"`
    Version      string              `json:"version"` // Schema version
}

// TeamSummary aggregates multiple users
type TeamSummary struct {
    TeamID       TeamID              `json:"team_id"`
    TimeRange    TimeRange           `json:"time_range"`
    Members      []UserSummary       `json:"members"`
    Aggregated   AggregatedTeamMetrics `json:"aggregated"`
    GeneratedAt  time.Time           `json:"generated_at"`
}

// AggregatedTeamMetrics for team-level insights
type AggregatedTeamMetrics struct {
    TotalMembers      int               `json:"total_members"`
    ActiveMembers     int               `json:"active_members"`
    
    // Aggregate counts
    TotalMessages     int               `json:"total_messages"`
    TotalCommits      int               `json:"total_commits"`
    TotalPRs          int               `json:"total_prs"`
    TotalIssues       int               `json:"total_issues"`
    TotalMeetings     int               `json:"total_meetings"`
    
    // Team scores
    TeamVelocity      float64           `json:"team_velocity"`
    CollaborationScore float64          `json:"collaboration_score"`
    CodeQualityAvg    float64           `json:"code_quality_avg"`
    
    // Top contributors
    TopContributors   []TopContributor  `json:"top_contributors"`
}

// TopContributor for leaderboard
type TopContributor struct {
    UserID    UserID  `json:"user_id"`
    Name      string  `json:"name"`
    Score     float64 `json:"score"`
    Breakdown map[string]int `json:"breakdown"` // {"commits": 50, "prs": 10}
}

// Metadata for versioning and timestamps
type Metadata struct {
    Version     string    `json:"version"`
    SchemaVersion string  `json:"schema_version"`
    GeneratedAt time.Time `json:"generated_at"`
    GeneratedBy string    `json:"generated_by"` // Service/user that generated
}
```

---

## 📊 **LAYER 10: HTTP & JSON TYPES**

```go
// ============================================================================
// HTTP LAYER
// ============================================================================

// HTTPRequest wraps http.Request for applicative processing
type HTTPRequest struct {
    Method      string            `json:"method"`
    Path        string            `json:"path"`
    RouteParams map[string]string `json:"route_params"`
    QueryParams map[string]string `json:"query_params"`
    Headers     map[string]string `json:"headers"`
    Body        []byte            `json:"body"`
}

// HTTPResponse for applicative response building
type HTTPResponse struct {
    StatusCode int               `json:"status_code"`
    Headers    map[string]string `json:"headers"`
    Body       []byte            `json:"body"`
}

// HTTPError for HTTP-level errors
type HTTPError struct {
    StatusCode int       `json:"status_code"`
    Message    string    `json:"message"`
    Code       string    `json:"code"` // "invalid_request", "not_found", "internal_error"
    Details    []string  `json:"details,omitempty"`
    Timestamp  time.Time `json:"timestamp"`
}

// StatusCode enum
type StatusCode int

const (
    StatusOK                  StatusCode = 200
    StatusCreated             StatusCode = 201
    StatusAccepted            StatusCode = 202
    StatusBadRequest          StatusCode = 400
    StatusUnauthorized        StatusCode = 401
    StatusForbidden           StatusCode = 403
    StatusNotFound            StatusCode = 404
    StatusTooManyRequests     StatusCode = 429
    StatusInternalServerError StatusCode = 500
    StatusServiceUnavailable  StatusCode = 503
)

// RouteParams extracted from URL path
type RouteParams struct {
    UserID string `json:"user_id,omitempty"`
    TeamID string `json:"team_id,omitempty"`
}

// QueryParams extracted from query string
type QueryParams struct {
    From      *time.Time `json:"from,omitempty"`
    To        *time.Time `json:"to,omitempty"`
    Audience  string     `json:"audience,omitempty"` // For LLM: "executive" | "technical" | "team"
}

// EncodingError for JSON encoding failures
type EncodingError struct {
    Field   string `json:"field"`
    Type    string `json:"type"`
    Message string `json:"message"`
}

// DecodingError for JSON decoding failures
type DecodingError struct {
    Field   string `json:"field"`
    Value   string `json:"value"`
    Message string `json:"message"`
}
```

---

## ✅ **Complete Domain Type Summary**

### **Total Custom Types Defined:**
- **Configuration:** 12 types
- **Platform Data (Raw):** 60+ types (Slack: 10, GitHub: 20, Jira: 25, Zoom: 15)
- **Normalization:** 6 types
- **Validation & Planning:** 4 types
- **Database:** 2 types
- **Analytics:** 1 type (+ many fields)
- **Correlation:** 3 types
- **LLM:** 4 types
- **Final Aggregation:** 5 types
- **HTTP & JSON:** 9 types

### **Reuse from purekernels:**
- ✅ All generic monoids (List, Sum, Avg, Max, Min)
- ✅ All applicatives (Concurrent, Validation, Reader, Writer, ZipList, Const)
- ✅ Fold/Reduce operations

---

## 🚀 **Next: Implementation with purekernels**

Now that we have complete, research-based domain types, we can implement:

1. **Custom monoids** for domain types (UserActivityMonoid, FetchResultMonoid, etc.)
2. **Connector implementations** using platform-specific SDKs
3. **Normalization functions** (Raw → Event)
4. **Applicative compositions** using purekernels
5. **REST API handlers** with Reader/Writer

**Ready for implementation?** 🎯



