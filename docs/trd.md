# Voice AI Customer Support Platform — Technical Requirements Document (TRD)

## Status
Draft — Pending Final Approval

## Version
1.1

## Author
Sreyash Reddy (IAmCyphr)

## Last Updated
2026-05-28

---

# Table of Contents

1. [Overview](#1-overview)
2. [System Architecture](#2-system-architecture)
3. [API Design](#3-api-design)
4. [Data Models](#4-data-models)
5. [Voice Pipeline](#5-voice-pipeline)
6. [RAG Pipeline](#6-rag-pipeline)
7. [MCP Integration](#7-mcp-integration)
8. [Session Management](#8-session-management)
9. [Multi-Tenant Isolation](#9-multi-tenant-isolation)
10. [Authentication & Authorization](#10-authentication--authorization)
11. [Role Management](#11-role-management)
12. [Permission Matrix](#12-permission-matrix)
13. [Error Handling & Resilience](#13-error-handling--resilience)
14. [Credit & Billing Management](#14-credit--billing-management)
15. [Notification System](#15-notification-system)
16. [Analytics & Logging](#16-analytics--logging)
17. [Infrastructure & Deployment](#17-infrastructure--deployment)
18. [Non-Functional Requirements](#18-non-functional-requirements)
19. [Real-Time Conversation Scoring](#19-real-time-conversation-scoring)
20. [System Prompt Generation](#20-system-prompt-generation)

---

# 1. Overview

This document specifies the technical requirements for the Voice AI Customer Support Platform. It translates product requirements (PRD) into implementable technical specifications.

## 1.2 Scope

| In Scope | Out of Scope |
|----------|--------------|
| Real-time voice AI customer support (Web App) | Mobile App (v0.2+) |
| Text chat support (Web App) | URL scraping (v0.2+) |
| Multi-tenant architecture | Business-provided MCP (v0.2+) |
| RAG-based knowledge retrieval | SSO integration |
| MCP tool integration | Custom domains |
| Human escalation flow | Pre-built connectors |
| Analytics dashboard | |
| Role-based access control | |
| Platform MCP servers (shared) | |

## 1.3 Reference Documents

- PRD: `docs/prd.md`
- UX Design: `docs/ux-design-specification.md`

## 1.4 Definitions

| Term | Definition |
|------|------------|
| **Org** | Organization — the business customer using the platform |
| **Agent** | Human support agent who takes escalated calls |
| **Customer** | End user contacting support via voice or chat |
| **Session** | Single conversation (voice or chat) between customer and AI |
| **Turn** | One exchange: customer speaks, AI responds |
| **Platform MCP** | MCP servers provided by the platform, shared across all orgs |
| **Org MCP** | MCP servers configured by each organization for their specific tools |

## 1.5 Acronyms

| Acronym | Definition |
|---------|------------|
| API | Application Programming Interface |
| JWT | JSON Web Token |
| MCP | Model Context Protocol |
| NFR | Non-Functional Requirement |
| OCR | Optical Character Recognition |
| RAG | Retrieval-Augmented Generation |
| REST | Representational State Transfer |
| SLA | Service Level Agreement |
| TBD | To Be Determined |
| TLS | Transport Layer Security |
| TPM | Tokens Per Minute |
| UUID | Universally Unique Identifier |
| WebSocket | Bidirectional communication protocol |

---

# 2. System Architecture

## 2.1 High-Level System Topology

```
+------------------------------------------------------------------------------+
|                              CLIENT LAYER                                    |
|                                                                              |
|   +----------------+  +------------------+  +---------------------------+   |
|   |   Web App      |  |   Embed         |  |   Admin                  |   |
|   |   (Voice +     |  |   Widget        |  |   Dashboard              |   |
|   |    Chat)       |  |   (External)    |  |   (Org + Super)          |   |
|   +----------------+  +------------------+  +---------------------------+   |
|                                                                              |
|   Mobile App (Future — v0.2+)                                                |
+------------------------------------------------------------------------------+
           |                        |                         |
           +------------------------+-------------------------+
                                    |
                                    v
+------------------------------------------------------------------------------+
|                              API GATEWAY                                      |
|                                                                              |
|   +----------------------------------------------------------------------+  |
|   |  Rate Limiting  |  Auth (JWT)  |  Tenant Context  |  Logging       |  |
|   +----------------------------------------------------------------------+  |
+------------------------------------------------------------------------------+
                                    |
                                    v
+------------------------------------------------------------------------------+
|                             CORE SERVICES                                     |
|                                                                              |
|   +----------------+  +----------------+  +----------------+  +-------------+  |
|   |  Voice/Chat   |  |  RAG          |  |  MCP          |  |  Analytics  |  |
|   |  Service      |  |  Service      |  |  Service      |  |  Service    |  |
|   |               |  |               |  |               |  |             |  |
|   |  WebSocket    |  |  Retrieve     |  |  Tool Exec    |  |  Aggregate  |  |
|   |  Gemini Proxy |  |  Embed       |  |  Timeout      |  |  Dashboard  |  |
|   |  Audio Proc   |  |  Chunk       |  |  Retry       |  |  Export     |  |
|   +--------+-------+  +--------+------+  +--------+------+  +------+------+  |
|            |                      |                    |                    |       |
|            +----------------------+--------------------+--------------------+       |
|                                     |                                        |
+-------------------------------------+----------------------------------------+
                                      v
+------------------------------------------------------------------------------+
|                               DATA LAYER                                     |
|                                                                              |
|   +----------------+  +----------------+  +----------------+  +-------------+  |
|   |  PostgreSQL    |  |  Redis        |  |  pgvector     |  |  S3         |  |
|   |  (pgvector)   |  |               |  |  (Extension)  |  |  Compatible |  |
|   |               |  |               |  |               |  |             |  |
|   |  Users/Orgs   |  |  Sessions     |  |  Embeddings   |  |  Documents  |  |
|   |  Roles/Perms  |  |  Cache       |  |  (per org)   |  |  Parquet    |  |
|   |  Transcripts  |  |  Rate Limit   |  |               |  |  Media      |  |
|   +----------------+  +----------------+  +----------------+  +-------------+  |
|                                                                              |
|   +----------------------------------------------------------------------+  |
|   |                    API KEY POOL MANAGER                               |  |
|   |                                                                      |  |
|   |   +----------+  +----------+  +----------+  +----------+            |  |
|   |   |  Key A   |  |  Key B   |  |  Key C   |  |  Key D   |            |  |
|   |   |  (Pool)  |  |  (Pool)  |  |  (Pool)  |  |  Backup  |            |  |
|   |   +----------+  +----------+  +----------+  +----------+            |  |
|   |                                                                      |  |
|   |   Dynamic Round-Robin  |  Monitor: RPM | TPM | Sessions  | Alert     |  |
|   +----------------------------------------------------------------------+  |
+------------------------------------------------------------------------------+
                                      |
                                      v
                    +-----------------------------------------+
                    |           Gemini Live API                |
                    |   WebSocket  |  STT  |  LLM  |  TTS   |
                    +-----------------------------------------+
```

## 2.2 Component Responsibilities

| Component | Responsibility |
|-----------|----------------|
| **Web App** | Browser-based voice/chat interface, audio capture/playback |
| **Embed Widget** | JavaScript widget for external websites |
| **Admin Dashboard** | Org management, documents, analytics, settings |
| **API Gateway** | Auth validation, tenant isolation, rate limiting, logging |
| **Voice/Chat Service** | WebSocket handling, Gemini proxy, audio processing |
| **RAG Service** | Document chunking, embedding, retrieval |
| **MCP Service** | Tool definitions, execution, timeout/retry, caching |
| **Analytics Service** | Data aggregation, dashboard queries |
| **API Key Pool Manager** | Key rotation, monitoring, failover |
| **PostgreSQL** | Primary data store (users, orgs, transcripts, embeddings) |
| **Redis** | Session state, caching, rate limiting |
| **pgvector** | Embedding storage and retrieval (PostgreSQL extension) |
| **S3 Compatible** | Document storage, Parquet exports (Phase 2) |

**Note:** Mobile App is planned for v0.2+ — not in MVP scope.

## 2.3 Data Flow: Voice Call

```
+--------+    +----------+    +----------------+    +--------------+
| Browser |-->| WebSocket |-->| Voice Service  |-->| Gemini Live  |
|  (Mic) |    | Connect  |    |                |    |    API       |
+--------+    +----+-----+    +--------+-------+    +--------------+
                 |                    |                    |
                 |  1. Auth          |  2. RAG Retrieve  |  3. Tool Call
                 |  2. org_id       |  3. MCP Config    | 4. Response
                 v                    v                    v
           +---------+          +---------+          +--------+
           |  Redis  |          | Vector DB |          |   MCP    |
           | Session  |          |   (RAG)   |          | Service  |
           +---------+          +---------+          +--------+
                                       |
                                       v
                               +--------------+
                               |   Response   |
                               |   (Audio +   |
                               |    Text)     |
                               +------+-------+
                                      |
+--------+    +----------+    +---------+-----+
| Browser |<--| WebSocket |<--| Voice Service |  (Bidirectional streaming)
|(Speaker|<--| Response  |<--|              |
+--------+    +----------+    +--------------+
```

## 2.4 Tech Stack Summary

| Category | Technology | Notes |
|----------|------------|-------|
| **Backend** | Go | Primary language |
| **Web Framework** | Fiber or Gin | With gorilla/websocket |
| **Primary DB** | PostgreSQL | With pgvector extension |
| **Session/Cache** | Redis | Sessions, caching, rate limits |
| **Vector DB** | pgvector | PostgreSQL extension, MVP choice |
| **Document Storage** | S3 Compatible | MinIO/R2/Backblaze/AWS S3 |
| **Voice AI** | Gemini Live API | Go SDK: `google.golang.org/genai` |
| **Tool Integration** | MCP | Model Context Protocol |
| **Analytics** | PostgreSQL (MVP), DuckDB (Phase 2) | |
| **Deployment** | Docker + Kamal | Self-hosted |

---

# 3. API Design

## 3.1 API Overview

| API Type | Protocol | Purpose |
|----------|----------|---------|
| **REST API** | HTTP/JSON | CRUD operations, admin functions |
| **WebSocket** | WSS | Real-time voice/chat streaming |

## 3.2 REST API Endpoints

### Authentication

```
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
```

### Organizations

```
GET    /api/v1/orgs                    # List orgs (Super Admin)
POST   /api/v1/orgs                    # Create org (Super Admin)
GET    /api/v1/orgs/:org_id            # Get org details
PATCH  /api/v1/orgs/:org_id            # Update org
DELETE /api/v1/orgs/:org_id            # Deactivate org (Super Admin)
```

### Users

```
GET    /api/v1/orgs/:org_id/users                    # List users
POST   /api/v1/orgs/:org_id/users                    # Create user
GET    /api/v1/orgs/:org_id/users/:user_id          # Get user
PATCH  /api/v1/orgs/:org_id/users/:user_id          # Update user
DELETE /api/v1/orgs/:org_id/users/:user_id          # Remove user
POST   /api/v1/orgs/:org_id/users/:user_id/roles    # Assign role
```

### Sessions

```
POST   /api/v1/sessions                        # Create session (voice or chat)
GET    /api/v1/sessions/:session_id            # Get session details
PATCH  /api/v1/sessions/:session_id            # Update session (escalate, end)
GET    /api/v1/sessions/:session_id/transcript  # Get transcript
```

### Documents

```
GET    /api/v1/orgs/:org_id/documents             # List documents
POST   /api/v1/orgs/:org_id/documents              # Upload document
GET    /api/v1/orgs/:org_id/documents/:doc_id      # Get document
DELETE /api/v1/orgs/:org_id/documents/:doc_id      # Delete document
PATCH  /api/v1/orgs/:org_id/documents/:doc_id      # Update metadata
```

### MCP Servers

```
GET    /api/v1/orgs/:org_id/mcp-servers                    # List MCP servers
POST   /api/v1/orgs/:org_id/mcp-servers                    # Add MCP server
GET    /api/v1/orgs/:org_id/mcp-servers/:server_id         # Get MCP server
PATCH  /api/v1/orgs/:org_id/mcp-servers/:server_id         # Update MCP server
DELETE /api/v1/orgs/:org_id/mcp-servers/:server_id         # Remove MCP server
GET    /api/v1/orgs/:org_id/mcp-servers/:server_id/tools   # List tools
POST   /api/v1/orgs/:org_id/mcp-servers/:server_id/test    # Test connection
```

### System Prompt

```
GET    /api/v1/orgs/:org_id/system-prompt      # Get system prompt
PATCH  /api/v1/orgs/:org_id/system-prompt      # Update greeting (org config)
```

### Analytics

```
GET    /api/v1/orgs/:org_id/analytics/overview         # Dashboard overview
GET    /api/v1/orgs/:org_id/analytics/conversations     # Conversation list
GET    /api/v1/orgs/:org_id/analytics/knowledge-gaps     # Knowledge gaps
GET    /api/v1/orgs/:org_id/analytics/trends            # Trend data
GET    /api/v1/orgs/:org_id/analytics/export            # Export report
```

### Notifications

```
GET    /api/v1/users/:user_id/notifications                      # List notifications
PATCH  /api/v1/users/:user_id/notifications/:notif_id           # Mark read
DELETE /api/v1/users/:user_id/notifications/:notif_id           # Dismiss
```

### Usage

```
GET    /api/v1/orgs/:org_id/usage          # Current usage (RPM, TPM)
```

## 3.3 WebSocket API

### Endpoint

```
WSS /api/v1/sessions/:session_id/stream
```

### Authentication

WebSocket authenticates via `Sec-WebSocket-Protocol` header:
```
Sec-WebSocket-Protocol: Bearer <jwt>
```

For v0.1 compatibility, query parameter `?token=<jwt>` is also supported but deprecated.
Tokens are masked in all server logs.

### Message Types

#### Client to Server

```json
// Audio chunk (voice)
{
  "type": "audio",
  "data": "<base64-encoded PCM16>",
  "mimeType": "audio/pcm"
}

// Text message (chat)
{
  "type": "text",
  "data": "Hello, I need help with my order"
}

// Tool response
{
  "type": "tool_response",
  "toolCallId": "call_123",
  "response": { "result": "..." }
}

// Interrupt (customer interrupting AI)
{
  "type": "interrupt"
}

// End session
{
  "type": "end"
}
```

#### Server to Client

```json
// Audio response
{
  "type": "audio",
  "data": "<base64-encoded PCM16>",
  "mimeType": "audio/pcm"
}

// Text response
{
  "type": "text",
  "data": "I can help you track your order."
}

// Transcription
{
  "type": "transcription",
  "role": "customer",
  "text": "I need to track my order",
  "finished": true
}

// Tool call
{
  "type": "tool_call",
  "toolCallId": "call_123",
  "tool": "lookup_order",
  "params": { "orderId": "12345" }
}

// Session state
{
  "type": "state",
  "state": "listening" | "thinking" | "speaking" | "interrupted" | "escalating" | "connecting"
}

// Error
{
  "type": "error",
  "code": "rate_limit" | "timeout" | "service_unavailable" | "payload_too_large",
  "message": "...",
  "requestId": "req_abc123",
  "retryAfterMs": 5000
}

// Buffer clear (on interrupt)
{
  "type": "buffer_clear"
}

// End
{
  "type": "end",
  "reason": "customer_hangup" | "escalated" | "completed"
}
```

### Session Lifecycle

```
Client                          Server                          Gemini
  |                                |                               |
  |  1. Connect (token)          |                               |
  |------------------------------->|                               |
  |                                |  2. Validate + Setup          |
  |                                |  3. MCP Tool Discovery       |
  |                                |  4. Build System Prompt      |
  |                                |------------------------------>|
  |                                |<------------------------------|
  |                                |      SETUP_COMPLETE           |
  |<-------------------------------|                               |
  |     CONNECTED                   |                               |
  |                                |                               |
  |  5. Audio chunk                |                               |
  |------------------------------->|  6. Forward audio            |
  |                                |------------------------------>|
  |                                |<------------------------------|
  |                                |      Audio + Text             |
  |<-------------------------------|                               |
  |     Response                    |                               |
  |                                |                               |
  |  [Repeat for conversation]     |                               |
  |                                |                               |
  |  7. Tool call (Gemini)        |                               |
  |<-------------------------------|                               |
  |  8. Tool response             |                               |
  |------------------------------->|  9. Forward to Gemini        |
  |                                |------------------------------>|
  |                                |<------------------------------|
  |<-------------------------------|                               |
  |     Continue                    |                               |
  |                                |                               |
  |  10. End session              |                               |
  |------------------------------->|  11. Close Gemini session     |
  |                                |------------------------------>|
  |     CLOSED                     |                               |
```

---

# 4. Data Models

## 4.1 PostgreSQL Schema

### Organizations

```sql
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    status VARCHAR(20) DEFAULT 'active'
        CHECK (status IN ('active', 'suspended', 'inactive')),
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_orgs_slug ON organizations(slug);
CREATE INDEX idx_orgs_status ON organizations(status);
```

### Users

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    role VARCHAR(50) NOT NULL CHECK (role IN ('admin', 'agent', 'member', 'customer')),
    custom_role_id UUID REFERENCES custom_roles(id),
    settings JSONB DEFAULT '{}',
    last_login TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(org_id, email)
);

CREATE INDEX idx_users_org ON users(org_id);
CREATE INDEX idx_users_email ON users(email);
```

**Password Hashing:** Argon2id with parameters:
- Memory: 64 MB
- Iterations: 3
- Parallelism: 4
- Salt: 16 bytes, randomly generated, embedded in hash output
- Max password length: 128 bytes (reject longer with 400 BAD_REQUEST)

### Custom Roles

```sql
CREATE TABLE custom_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    permissions JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(org_id, name)
);

CREATE INDEX idx_custom_roles_org ON custom_roles(org_id);
```

### Sessions

```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    customer_id VARCHAR(255),  -- External customer identifier (org-provided)
    mode VARCHAR(20) NOT NULL CHECK (mode IN ('voice', 'chat')),
    status VARCHAR(20) DEFAULT 'active'
        CHECK (status IN ('active', 'completed', 'escalated', 'failed')),
    resolution_status VARCHAR(20)
        CHECK (resolution_status IN ('resolved_ai', 'resolved_human', 'unresolved', 'unknown')),
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE,
    satisfaction_score INTEGER CHECK (satisfaction_score BETWEEN 1 AND 5),
    escalation_reason VARCHAR(100),
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_sessions_org ON sessions(org_id);
CREATE INDEX idx_sessions_status ON sessions(status);
CREATE INDEX idx_sessions_started ON sessions(started_at);
CREATE INDEX idx_sessions_customer ON sessions(org_id, customer_id);
```

### Conversations (Transcripts)

```sql
CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('customer', 'ai', 'agent')),
    content TEXT NOT NULL,
    audio_url VARCHAR(500),  -- Nullable. Populated only if org has audio storage enabled.
    tool_calls JSONB,
    turn_index INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_conversations_session ON conversations(session_id);
CREATE INDEX idx_conversations_session_turn ON conversations(session_id, turn_index);
```

### Documents

```sql
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    file_type VARCHAR(50) NOT NULL CHECK (file_type IN ('pdf', 'docx', 'txt', 'md', 'html')),
    file_path VARCHAR(500) NOT NULL,
    file_size INTEGER,
    category VARCHAR(100),
    status VARCHAR(20) DEFAULT 'processing'
        CHECK (status IN ('processing', 'active', 'outdated')),
    chunks_count INTEGER,
    metadata JSONB DEFAULT '{}',
    uploaded_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_docs_org ON documents(org_id);
CREATE INDEX idx_docs_status ON documents(status);
```

### Document Embeddings (pgvector)

```sql
CREATE TABLE document_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    chunk_index INTEGER NOT NULL,
    content TEXT NOT NULL,
    embedding VECTOR(768),  -- Dimensionality depends on embedding model
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_embeddings_org ON document_embeddings(org_id);
CREATE INDEX idx_embeddings_doc ON document_embeddings(document_id);
CREATE INDEX idx_embeddings_vector ON document_embeddings USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);
```

### MCP Servers

```sql
CREATE TABLE mcp_servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(500) NOT NULL,
    auth_type VARCHAR(50) CHECK (auth_type IN ('none', 'api_key', 'oauth')),
    auth_config_encrypted BYTEA,  -- AES-256-GCM encrypted credentials
    status VARCHAR(20) DEFAULT 'disconnected'
        CHECK (status IN ('connected', 'disconnected', 'error')),
    server_type VARCHAR(20) DEFAULT 'org'
        CHECK (server_type IN ('platform', 'org')),
    last_tested TIMESTAMP WITH TIME ZONE,
    tools JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_mcp_org ON mcp_servers(org_id);
CREATE INDEX idx_mcp_type ON mcp_servers(server_type);
```

### Notifications

```sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT,
    data JSONB DEFAULT '{}',
    read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_notifs_user ON notifications(user_id);
CREATE INDEX idx_notifs_user_read ON notifications(user_id, read);
CREATE INDEX idx_notifs_created ON notifications(created_at DESC);
```

### Audit Logs

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    details JSONB DEFAULT '{}',
    ip_address VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_audit_org ON audit_logs(org_id);
CREATE INDEX idx_audit_user ON audit_logs(user_id);
CREATE INDEX idx_audit_created ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_org_user_time ON audit_logs(org_id, user_id, created_at DESC);
```

## 4.2 Redis Data Structures

### Session State

```
Key: session:{session_id}
Type: Hash
TTL: 30 minutes (refreshed on user activity)
Fields:
  - org_id: string (UUID)
  - user_id: string (UUID)
  - mode: voice|chat
  - status: active|escalated|completed
  - started_at: ISO8601 timestamp
  - last_activity: ISO8601 timestamp
  - conversation_history: JSON array (last 20 turns)
  - context_chunks: JSON array
  - mcp_tool_state: JSON object
  - current_turn: integer
  - preserved_entities: JSON array (order_*, account_*, ticket_* patterns)
```

### Conversation History

```
Key: session:{session_id}:history
Type: List
TTL: 60 minutes
Fields: JSON objects [{role, content, timestamp, tool_calls}]
```

### Rate Limiting

```
Key: ratelimit:org:{org_id}
Type: String (counter)
TTL: 60 seconds
Value: Request count
```

### API Key Pool State

```
Key: apikey:global:pool
Type: Sorted Set
TTL: None (persistent)
Members: Key hashes
Scores: Last used timestamp

Key: apikey:key:{key_hash}:state
Type: Hash
TTL: None (persistent)
Fields:
  - status: active|backup|failed
  - current_rpm: integer
  - current_tpm: integer
  - concurrent_sessions: integer
  - last_used: timestamp

Key: apikey:key:{key_hash}:failover_lock
Type: String
TTL: 30 seconds (auto-release)
Value: "1" (distributed lock for failover)
```

### MCP Tool Cache

```
Key: mcp:tools:{server_id}
Type: String (JSON)
TTL: 5 minutes
Value: Tool definitions array

Key: mcp:platform:tools
Type: String (JSON)
TTL: 30 minutes (refreshed on service start)
Value: Platform-wide tool definitions
```

### User Notifications

```
Key: notifications:unread:{user_id}
Type: String (integer)
Value: Count of unread notifications
```

## 4.3 S3 Path Structure

```
s3://bucket/
├── orgs/
│   └── {org_id}/
│       ├── documents/
│       │   └── {document_id}/
│       │       └── original.{ext}
│       ├── sessions/
│       │   └── {session_id}/
│       │       └── audio.pcm
│       └── exports/
│           └── {date}/
│               └── analytics.parquet
```

### Audio Storage Lifecycle

**During a session:**
- Audio chunks are written to local temp disk at `/tmp/sessions/{session_id}/audio.pcm`
- Audio is never stored in Redis
- Each chunk is appended to the file as raw PCM

**Post-session (worker):**
- Worker uploads audio file from temp disk to S3 asynchronously
- Path: `orgs/{org_id}/sessions/{session_id}/audio.pcm`
- On success: `conversations.audio_url` updated for all turns from that session
- Temp file deleted after 24-hour grace period post-upload

**Retention:**
- Default: Auto-delete from S3 after 30 days
- Configurable per org: Super Admin can set longer retention up to 1 year

---

# 5. Voice Pipeline

## 5.1 Audio Flow

```
+------------+    +------------+    +------------+    +------------+
|  Browser   |    |   Voice    |    |   Gemini    |    |   Voice    |
| Microphone |-->|  Service   |-->|  Live API   |<--|  Service   |
|            |    |            |    |            |    |            |
|            |    | 1. Auth   |    | 2. Stream   |    | 3. Audio  |
|            |    | 2. Buffer |    |   Audio     |    | 4. Forward|
|            |    | 3. Forward |    | 3. STT      |    |           |
|            |    |            |    | 4. LLM       |    |           |
|            |    |            |    | 5. TTS       |    |           |
+------------+    +------------+    +------------+    +------------+
                                                           |
                                                           v
                                                     +------------+
                                                     |  Browser    |
                                                     |  Speaker    |
                                                     +------------+
```

## 5.2 Audio Specifications

| Parameter | Input (Mic) | Output (Speaker) | Storage |
|-----------|-------------|-----------------|---------|
| **Format** | PCM 16-bit | PCM 16-bit | PCM 16-bit |
| **Sample Rate** | 16,000 Hz | 16,000 Hz (resampled from 24kHz) | 16,000 Hz |
| **Bit Depth** | 16-bit | 16-bit | 16-bit |
| **Channels** | Mono | Mono | Mono |
| **Endianness** | Little-endian | Little-endian | Little-endian |
| **Mime Type** | `audio/pcm` | `audio/pcm` | `audio/pcm` |
| **Chunk Duration** | ~32ms (512 samples) | ~32ms | N/A |

**Notes:**
- Gemini Live API outputs 24kHz audio; backend resamples to 16kHz before storage/transmission
- Browser handles 24kHz playback in its own AudioContext
- Max chunk size: 64KB (reject larger with `PAYLOAD_TOO_LARGE` error)

## 5.3 Session Initialization Sequence

```
1. Client initiates WebSocket connection with JWT
2. Backend validates JWT and extracts org_id
3. Backend fetches session state from Redis
4. Backend retrieves MCP server configs for org
5. Backend fetches cached platform MCP tools (pre-warmed)
6. Backend connects to org MCP servers (on-demand, sequential)
7. Backend discovers tools, validates schemas
8. Backend builds system prompt with:
   - Base system instruction (platform)
   - Greeting message (org-configurable)
   - RAG context (if available)
   - Platform MCP tools
   - Org MCP tools
9. Backend establishes Gemini Live session with full config
10. Backend sends CONNECTED state to client
11. Client begins streaming audio

Target initialization time: < 2 seconds (with pre-warmed MCP cache)
```

## 5.4 Gemini Live Output Capture

Per turn, Gemini Live returns three outputs through the same WebSocket stream. All three are captured and stored.

| Output | Type | Description |
|--------|------|-------------|
| **Customer Transcription** | `{"type": "transcription", "role": "customer", ...}` | STT result of customer speech |
| **AI Text Response** | `{"type": "text", "data": "..."}` | AI's text reply |
| **AI Audio Response** | `{"type": "audio", "data": "<base64>", ...}` | AI's spoken response (TTS) |

Storage behavior:
- **Transcription + AI text:** Stored in `conversations` table (per turn, one row per speaker)
- **AI audio:** Written to temp disk `/tmp/sessions/{session_id}/audio.pcm`, uploaded to S3 post-session
- **Customer audio:** Written to temp disk alongside AI audio (interleaved or separate), uploaded to S3 post-session

**Note:** All three streams arrive on the same WebSocket connection but may arrive at slightly different times. Each is tagged with sequence metadata to allow accurate interleaving.

## 5.5 Context Management

### Context Window

Gemini Live API context window limit: ~10-15 minutes of conversation.

**Sliding Window Compression:**
- After 20 turns, oldest turns are compressed
- Last 20 turns + system prompt always retained
- Entity preservation: tokens matching patterns `order_*`, `account_*`, `ticket_*`, `customer_*` are never discarded

```
Session Start
    |
    v
[Turn 1] [Turn 2] [Turn 3] ... [Turn 20]
    |
    v
If turns > 20:
    |
    v
Compress oldest turns
Keep: Last 20 turns + system prompt + preserved entities
Discard: Filler content before
    |
    v
Continue conversation
```

### Session Resumption

If WebSocket connection drops:
1. Client reconnects with same session_id
2. Backend validates session from Redis
3. Backend checks conversation_history in Redis
4. Backend rebuilds Gemini config with history
5. Backend reconnects to Gemini with context
6. Backend sends RESUMED state to client

**Requirements:**
- Redis TTL must cover max expected disconnection window
- Conversation history replay limited to last 20 turns
- Preserved entities always included

## 5.6 Audio Interruption Handling

When client sends `{"type": "interrupt"}`:

```
1. Backend sends {"type": "buffer_clear"} to client
2. Backend flushes any pending audio in Redis buffer
3. Backend sends stop signal to Gemini (cancel generation)
4. Backend updates session state to "interrupted"
5. Backend waits for next customer input
```

## 5.7 Voice States

| State | Description | Client UI |
|-------|-------------|-----------|
| **idle** | Waiting for customer | Avatar static |
| **connecting** | Session initializing | Loading indicator |
| **listening** | Capturing customer speech | Waveform active |
| **thinking** | Processing, RAG retrieval, MCP calls | Animated dots |
| **speaking** | AI responding | Avatar animated |
| **interrupted** | Customer interrupted | Brief indicator |
| **escalating** | Transferring to human | Connecting state + wait time |
| **ended** | Session complete | Satisfaction prompt |

## 5.8 WebSocket Keepalive

```
Ping interval: 30 seconds
Pong timeout: 10 seconds
On pong timeout: Close connection, mark session as failed
```

---

# 6. RAG Pipeline

## 6.1 Document Ingestion Flow

```
+------------+    +------------+    +------------+    +------------+
| Document   |-->| Validation |-->|  Chunking  |-->| Embedding  |
| Upload     |    | (type,    |    | (semantic) |    | (batch)    |
| (S3)      |    |  size)    |    | 512 tokens |    |            |
+------------+    +------------+    +------------+    +------------+
                                                        |
                                                        v
                                                  +------------+
                                                  |  pgvector   |
                                                  |  (Indexed)  |
                                                  +------------+
```

## 6.2 Chunking Strategy

| Parameter | Value | Rationale |
|-----------|-------|-----------|
| **Chunk Size** | 512 tokens | Optimal for semantic search |
| **Chunk Overlap** | 64 tokens | Prevents boundary concept loss |
| **Method** | Semantic (sentence-aware) | Prefer splitting on sentence boundaries |
| **Min Size** | 100 tokens | Discard chunks below threshold |
| **Max Size** | 1024 tokens | Split oversized chunks |

## 6.3 Retrieval Flow

```
Query
  |
  v
+------------------+
| Intent           |
| Classification   |
+------------------+
  |
  v
+------------------+
| Embed Query       |
+------------------+
  |
  v
+------------------+
| Vector Search     |
| (Top 20 results) |
+------------------+
  |
  v
+-------------+-------------+
| Score > 0.75 | Score < 0.75 |
| Use        | Discard       |
+-------------+-------------+
  |
  v
+------------------+
| Top-K Chunks      |
| (max 10 chunks)    |
| + Context Header  |
+------------------+
  |
  v
+------------------+
| Build Prompt      |
| + Inject Context |
+------------------+
  |
  v
+------------------+
| Send to Gemini    |
+------------------+
```

## 6.4 Relevance Gate

| Score Range | Action |
|-------------|--------|
| **> 0.75** | Include — confident response |
| **0.50 - 0.75** | Include but flag uncertainty in response |
| **< 0.50** | Exclude — escalate or return "I don't have information" |

**"No Relevant Context" Path:**
- All retrieved chunks score < 0.50
- Agent triggers escalation or returns fallback message
- Conversation flagged for knowledge gap analysis

---

# 7. MCP Integration

## 7.1 MCP Server Types

### Platform MCP Servers
- Shared across all organizations
- Pre-warmed on service startup
- Tools cached in Redis (30 min TTL)
- Examples: generic_order_lookup, create_ticket, knowledge_search

### Org MCP Servers
- Configured by each organization
- Connected on-demand per session
- Credentials encrypted (AES-256-GCM)
- Examples: CRM integrations, proprietary APIs

## 7.2 Tool Calling Flow

```
+--------+    +-----------+    +-----------+    +----------+
| Gemini |-->|  Voice    |-->|   MCP     |-->| External  |
| (Call) |    | Service  |    |  Service  |    | Server    |
|        |    |          |    |           |    |           |
|lookup_ |    | Validate |    | Execute   |    | API Call  |
|order   |    | + Route  |    | + Timeout |    |           |
+--------+    +-----------+    +-----------+    +----------+
                                       |
                                       v
                                  +----------+
                                  | Response |
                                  | + Return |
                                  +----------+
                                       |
                                       v
                                  +----------+
                                  |  Gemini  |
                                  |(Continue)|
                                  +----------+
```

## 7.3 Timeout & Retry Configuration

| Scenario | Timeout | Retry |
|----------|--------|-------|
| **MCP Request** | 5 seconds | 2 retries with exponential backoff (1s, 2s) |
| **MCP Response > 1MB** | Reject | Log warning |
| **MCP Unavailable** | Fail fast | Continue with RAG-only |

## 7.4 Security Controls

| Control | Implementation |
|---------|----------------|
| **TLS** | TLS 1.3 mandatory for all MCP connections |
| **Certificate Pinning** | SHA-256 fingerprint validation |
| **API Key Auth** | Passed via `X-API-Key` header, never in URL |
| **IP Allowlist** | Only configured URLs callable; resolved on connection |
| **DNS Rebinding Protection** | Re-resolve on every connection; pin IP; reject redirects |
| **Request Size Limits** | Max 64KB request, 1MB response |
| **Network Isolation** | MCP calls from isolated network segment |
| **Method Restrictions** | GET/POST only unless explicitly configured |
| **Input Validation** | Strict JSON schema validation before execution |
| **Execution Timeout** | Max 5 seconds per tool call |

## 7.5 Credential Storage

MCP credentials encrypted with AES-256-GCM:
- Encryption key stored in environment variable or secrets manager
- Key derivation: PBKDF2 from secrets manager
- Salt: Unique per credential

---

# 8. Session Management

## 8.1 Session Lifecycle

```
+--------------------------------------------------------------------------+
|                         SESSION LIFECYCLE                                |
|                                                                          |
|   +--------+     +--------+     +-----------+     +-------+     +----+  |
|   | START  |---->| ACTIVE |---->| ESCALATE  |---->|  END  |     |FAIL|  |
|   |        |     |        |     | (optional)|     |       |     |    |  |
|   +--------+     +---+----+     +-----+-----+     +---+---+     +----+  |
|                      |                    |             |                  |
|                      | Error/Timeout      |             |                  |
|                      v                    v             v                  |
|                 +---------+        +---------+    +---------+            |
|                 | FAILED  |        |AGENT WAIT|    | COMPLETE|         |
|                 +---------+        +-----+-----+    +---------+            |
|                                    |              |                      |
|                                    | No agent      |                      |
|                                    | available     |                      |
|                                    v              |                      |
|                              +-----------+        |                      |
|                              | FALLBACK  |        |                      |
|                              | (Email +  |        |                      |
|                              |  Notify)  |        |                      |
|                              +-----------+        |                      |
+--------------------------------------------------------------------------+
```

## 8.2 Escalation Flow

### Escalation Trigger Conditions
- Customer explicitly requests human ("talk to agent")
- AI confidence below threshold after retries
- MCP call fails repeatedly
- Customer asks about unsupported topic

### Escalation Sequence
```
1. Customer requests escalation
    |
    v
2. AI acknowledges: "Let me connect you with a support specialist"
    |
    v
3. Backend:
    a. Gracefully closes Gemini Live session
    b. Persists conversation to PostgreSQL immediately
    c. Creates escalation record with full context
    d. Notifies available agents (in-app + configured channels)
    e. Updates session status to "escalated"
    |
    v
4. Agent receives:
    - Customer context (name, issue summary)
    - Full conversation transcript (read-only)
    - Knowledge gaps flagged by AI
    - Session duration, satisfaction hints
    |
    v
5. If no agent available:
    a. Send notification to org (email/in-app)
    b. Customer sees: estimated wait time or callback option
    c. Offer: Email | Message | Callback (org-configured)
```

## 8.3 Redis Session Schema

```json
{
  "id": "uuid",
  "org_id": "uuid",
  "user_id": "uuid",
  "mode": "voice",
  "status": "active",
  "started_at": "2026-05-28T10:00:00Z",
  "last_activity": "2026-05-28T10:05:00Z",
  "conversation_history": [
    {"role": "customer", "content": "I need help", "timestamp": "..."},
    {"role": "ai", "content": "How can I help?", "timestamp": "..."}
  ],
  "context_chunks": ["chunk_1", "chunk_2"],
  "mcp_tool_state": {"last_tool": "lookup_order", "result": "..."},
  "preserved_entities": [
    {"type": "order_id", "value": "ORD-12345"},
    {"type": "account_id", "value": "ACC-67890"}
  ],
  "metadata": {
    "language": "en",
    "voice_name": "Puck"
  }
}
```

## 8.4 Session TTL Policy

| Event | TTL Action |
|-------|------------|
| **Session created** | 30 minutes from start |
| **Activity (user message)** | Reset to 30 minutes |
| **Activity (AI response)** | No reset |
| **Session ends normally** | Delete after 1 hour |
| **Session fails/escalates** | Delete after 24 hours |

## 8.5 Idempotency

All WebSocket messages include an idempotency key:
```json
{
  "type": "audio",
  "data": "...",
  "id": "msg_uuid"  // Client-generated, unique per message
}
```

Backend deduplicates messages received within 5 seconds with same ID.

---

# 9. Multi-Tenant Isolation

## 9.1 Isolation Strategy

Soft multi-tenancy with strict logical enforcement.

Every request enforces org_id — no cross-org data access possible.

## 9.2 Isolation Layers

```
+------------------------------------------------------------------+
|                        API REQUEST                                |
+------------------------------------------------------------------+
                              |
                              v
+------------------------------------------------------------------+
|                      AUTH MIDDLEWARE                              |
|   +------------------------------------------------------------+ |
|   | Validate JWT signature                                     | |
|   | Extract user_id + org_id                                  | |
|   | Inject into request context                                | |
|   +------------------------------------------------------------+ |
+------------------------------------------------------------------+
                              |
                              v
+------------------------------------------------------------------+
|                      TENANT MIDDLEWARE                            |
|   +------------------------------------------------------------+ |
|   | Validate org_id from JWT matches request                   | |
|   | Enforce org_id on ALL database queries                     | |
|   | Enforce org_id on Redis operations                         | |
|   | Enforce org_id on Vector DB namespace (pgvector WHERE)      | |
|   | Enforce org_id on S3 paths                                 | |
|   +------------------------------------------------------------+ |
+------------------------------------------------------------------+
                              |
                              v
+------------------------------------------------------------------+
|                        RESPONSE                                   |
+------------------------------------------------------------------+
```

## 9.3 Threat Model & Mitigations

| Threat | Mitigation |
|--------|-----------|
| **JWT forged with different org_id** | Validate signature; verify org_id claim; reject |
| **SQL injection** | Parameterized queries; org_id enforced at query level |
| **Vector DB collision** | pgvector uses org_id in WHERE clause |
| **S3 prefix traversal** | S3 keys constructed server-side only; no user input in path |
| **MCP credential cross-org** | Credentials scoped by org_id; encrypted storage |
| **Redis data leakage** | Keys prefixed with org_id; validated on access |

## 9.4 Testing Requirements

- Unit tests: All queries include org_id filter
- Integration tests: Cross-org access attempts return 403
- Security audit: Quarterly review of data access patterns

---

# 10. Authentication & Authorization

## 10.1 Authentication Flow

```
+--------+    +----------+    +----------+    +----------+
|  User  |-->|  Login   |-->|  Verify  |-->|  Issue   |
|        |    | Request |    |  Creds  |    |   JWT   |
+--------+    +----------+    +----------+    +----------+
                                             |
                                             v
                                      +----------+
                                      |  Store   |
                                      |  Refresh |
                                      |  Token   |
                                      +----------+
```

## 10.2 JWT Structure

```json
{
  "sub": "user_uuid",
  "org_id": "org_uuid",
  "role": "admin",
  "custom_role_id": "custom_role_uuid",
  "permissions": ["docs:read", "docs:write", "analytics:read"],
  "exp": 1735689600,
  "iat": 1735686000
}
```

**JWT Requirements:**
- Algorithm: RS256 or HS256
- Expiry: Access token 15 minutes, Refresh token 7 days
- Refresh token stored server-side (Redis)

## 10.3 Authorization Flow

```
Request -> Auth Middleware -> Validate JWT -> Extract Permissions
                                        |
                                        v
                                  +-------------+
                                  | Check       |
                                  | Permission  |
                                  +------+------+
                                         |
                            +------------+------------+
                            v                         v
                       +--------+              +--------+
                       | Allow  |              | Deny   |
                       | Proceed|              |  403   |
                       +--------+              +--------+
```

---

# 11. Role Management

## 11.1 Role Hierarchy

```
PLATFORM LEVEL
|
+-- Super Admin (Write All)
|   - Full platform access
|   - Manage all orgs
|   - System settings
|   - Billing management
|
+-- Super Admin (Read All)
    - Read-only platform access
    - View all orgs
    - No billing/settings access

ORGANIZATION LEVEL
|
+-- Admin
|   - Full org access
|   - Create custom roles
|   - Assign permissions
|   - Configure MCP servers
|   - Billing access
|
+-- Agent
|   - Handle escalated calls
|   - Chat with customers
|   - View assigned conversations
|   - No billing/settings access
|
+-- Member (Custom Role)
|   - Configurable permissions
|   - Admin-created
|
+-- Customer
    - Voice/chat support only
    - No dashboard access
```

## 11.2 Fixed Roles

| Role | Level | Description |
|------|-------|-------------|
| **Super Admin (Write)** | Platform | Full access to all orgs, settings, billing |
| **Super Admin (Read)** | Platform | Read-only access to all orgs, settings |
| **Admin** | Org | Full access within org, create roles |
| **Agent** | Org | Handle escalations, view conversations |
| **Customer** | - | Voice/chat only, no dashboard |

## 11.3 Custom Roles

Admin creates custom roles with granular permission toggles.

Example: "Billing Viewer"
```
Custom Role: Billing Viewer
Permissions:
  - docs:read        [ON]
  - docs:write       [OFF]
  - conv:view         [ON]
  - conv:monitor      [OFF]
  - conv:takeover    [OFF]
  - analytics:view    [ON]
  - analytics:export  [ON]
  - settings:billing [ON]
  - settings:billing:write [OFF]
```

---

# 12. Permission Matrix

## 12.1 Permission Categories

| Category | Permissions |
|----------|-------------|
| **Documents** | docs:read, docs:write, docs:delete, docs:metadata |
| **Knowledge Base** | kb:configure, kb:chunks, kb:analytics |
| **AI / Voice** | ai:behavior, ai:mcp, ai:escalation, ai:prompts |
| **Team** | team:invite, team:permissions, team:remove, team:activity |
| **Conversations** | conv:view, conv:monitor, conv:takeover |
| **Analytics** | analytics:view, analytics:export |
| **Settings** | settings:profile, settings:billing, settings:webhooks |

## 12.2 Permission Toggle Matrix

| Permission | Admin | Agent | Custom (Example) |
|------------|-------|-------|-----------------|
| docs:read | Yes | Yes | Yes |
| docs:write | Yes | No | Configurable |
| docs:delete | Yes | No | No |
| conv:view | Yes | Yes | Yes |
| conv:monitor | Yes | No | Configurable |
| conv:takeover | Yes | Yes | No |
| team:manage | Yes | No | No |
| analytics:view | Yes | Yes | Yes |
| analytics:export | Yes | No | Configurable |
| settings:billing | Yes | No | No |
| settings:profile | Yes | Yes | Configurable |

---

# 13. Error Handling & Resilience

## 13.1 Error Categories

| Category | Examples | User Response |
|----------|----------|--------------|
| **Audio** | Mic denied, WebRTC failed | "Microphone access needed" |
| **Network** | Connection lost, timeout | Auto-retry with indicator |
| **RAG** | No context, vector DB unavailable | Escalate or generic response |
| **MCP** | Timeout, invalid response | Continue with RAG-only |
| **LLM** | Gemini unavailable, rate limit | "Technical difficulties" |
| **Auth** | Token expired, invalid org | "Session expired, refresh" |

## 13.2 WebSocket Error Schema

```json
{
  "type": "error",
  "code": "rate_limit" | "timeout" | "service_unavailable" | "payload_too_large",
  "message": "Human-readable description",
  "requestId": "req_abc123",
  "retryAfterMs": 5000
}
```

## 13.3 Resilience Patterns

### Retry + Backoff

| Attempt | Delay |
|---------|-------|
| 1 | 1 second |
| 2 | 2 seconds |
| 3 | 4 seconds |

### Circuit Breaker

| State | Condition | Behavior |
|-------|-----------|----------|
| **Closed** | < 5 failures | Normal operation |
| **Open** | 5 failures in 30s | Return fallback immediately |
| **Half-open** | 30s elapsed | Allow 1 test request |

### Fallback Chain

```
Primary API Key
    |
    +-- OK -> Normal
    |
    +-- Rate limited -> Retry with backoff
    |
    +-- Fails -> Secondary API Key
    |
    +-- Fails -> Alert + Queue
                        |
                        +-- Escalate to human
```

## 13.4 Graceful Degradation

| Level | Condition | User Experience |
|-------|-----------|-----------------|
| **Full** | All services available | Normal |
| **Degraded** | RAG unavailable | "I can help with general questions..." |
| **Degraded** | MCP unavailable | Continue with RAG only |
| **Critical** | Gemini unavailable | "Technical difficulties, please try again" |
| **Outage** | All down | Escalate immediately |

## 13.5 Connection Recovery

| Drop Duration | Action |
|--------------|--------|
| < 5 seconds | Auto-reconnect, preserve context |
| 5-30 seconds | Show "reconnecting" UI, retry 3x |
| > 30 seconds | End gracefully, flag incomplete |
| > 2 minutes | Mark for review, notify admin |

---

# 14. Credit & Billing Management

## 14.1 API Key Architecture

```
+--------------------------------------------------------------------------+
|                        API KEY POOL MANAGER                              |
|                                                                           |
|   +---------------------------------------------------------------------+ |
|   |  Key Allocation                                                      | |
|   |                                                                      | |
|   |   +----------+  +----------+  +----------+  +----------+           | |
|   |   |  Key A   |  |  Key B   |  |  Key C   |  |  Key D   |           | |
|   |   |  (Pool)  |  |  (Pool)  |  |  (Pool)  |  |  Backup  |           | |
|   |   +----------+  +----------+  +----------+  +----------+           | |
|   +---------------------------------------------------------------------+ |
|                                                                           |
|   +---------------------------------------------------------------------+ |
|   |  Load Distribution                                                     | |
|   |                                                                      | |
|   |   Dynamic round-robin across keys                                    | |
|   |   Skip keys approaching rate limit                                    | |
|   |   Monitor: RPM | TPM | Concurrent Sessions                           | |
|   +---------------------------------------------------------------------+ |
|                                                                           |
|   +---------------------------------------------------------------------+ |
|   |  Alert Thresholds                                                     | |
|   |                                                                      | |
|   |   > 50% used  -> Warning to Super Admin                             | |
|   |   > 80% used  -> Urgent alert to Super Admin                        | |
|   |   ~100% used  -> Failover to backup key                            | |
|   +---------------------------------------------------------------------+ |
+--------------------------------------------------------------------------+
```

## 14.2 Key Pool Configuration

Pool size is configurable. Default: 1 key per pool, pools grouped by org.

| Pool | API Key | Orgs | Purpose |
|------|---------|------|---------|
| Pool 1 | Key A | Configurable | Normal operations |
| Pool 2 | Key B | Configurable | Normal operations |
| Pool 3 | Key C | Configurable | Normal operations |
| Backup | Key D | All | Fallback when any pool key fails |

## 14.3 Monitoring

| Metric | Per Key | Per Pool | Per Org (Dashboard) |
|--------|---------|----------|---------------------|
| **RPM** | Yes | Yes | No (internal) |
| **TPM** | Yes | Yes | Yes (cost allocation) |
| **Concurrent Sessions** | Yes | Yes | No |
| **Token Cost** | Yes | Yes | Yes (calculated) |

## 14.4 Failover Behavior

```
Normal Operation
    |
    v
Key A handles its pool
    |
    v
Monitor: RPM | TPM | Sessions
    |
    +-- < 50% capacity -> Normal
    |
    +-- 50-80% capacity -> Alert Super Admin
    |   "Approaching limit, prepare backup"
    |
    +-- 80-95% capacity -> URGENT Alert
    |   "Add key now or risk outage"
    |
    +-- ~100% or failure -> Distributed lock acquired
        |
        v
        Key D (backup) takes over
        |
        v
        Alert: "Operating on backup key"
```

## 14.5 Cost Tracking

| Level | What We Track | Visible In |
|-------|--------------|------------|
| **Platform** | Total spend | Super Admin |
| **Pool** | Pool spend | Super Admin |
| **Org** | Org proportional cost | Org Admin (dashboard) |

**Calculation:**
```
Org Cost = (Org's TPM / Total Platform TPM) x Total Platform Spend
```

**Note:** This is an approximation. For precise billing, per-session token counts from Gemini usage logs are the source of truth.

---

# 15. Notification System

## 15.1 Notification Types

| Type | Trigger | Recipient | Channel |
|------|---------|-----------|---------|
| **Credit Warning** | Credits < 50% | Super Admin | In-app, Email |
| **Credit Critical** | Credits < 20% | Super Admin | In-app, Email, SMS |
| **Credit Exhausted** | Credits < 5% | Super Admin | In-app, Email, SMS, urgent |
| **Key Pool Warning** | Pool > 50% | Super Admin | In-app |
| **Key Pool Failure** | Pool key fails | Super Admin | In-app, Email |
| **MCP Server Down** | MCP unreachable | Org Admin | In-app, Email |
| **Escalation Threshold** | X escalations/hour | Org Admin | In-app, Email |
| **System Health** | Service unhealthy | Super Admin | In-app, Email |
| **Document Processed** | Doc indexing complete | Org Admin | In-app |

## 15.2 Notification Preferences

| Channel | Configurable | Options |
|---------|-------------|---------|
| **In-app** | Always on | Unread badge |
| **Email** | Yes | On/Off, digest frequency |
| **SMS** | Yes | On/Off for critical only |
| **Teams/Slack** | Yes | Webhook URL |

## 15.3 Escalation Fallback

When no human agent is available during escalation:

1. Customer sees: Estimated wait time or callback option
2. Org receives notification via configured channels
3. Customer can select: Email | Message | Callback

---

# 16. Analytics & Logging

## 16.1 Analytics Data Points

| Category | Metrics |
|----------|---------|
| **Volume** | Total calls/chats, daily/weekly/monthly |
| **Escalation** | % escalated, reasons, resolution status |
| **AI Performance** | Answer success rate, struggle topics, knowledge gaps |
| **Knowledge Gaps** | Unanswered questions flagged |
| **Satisfaction** | Feedback, ratings (1-5), implicit resolution (call-back) |
| **Additional** | Peak hours, failed queries, topic breakdown |

## 16.2 Resolution Tracking

| Status | Definition |
|--------|------------|
| **resolved_ai** | AI handled without escalation, customer satisfied |
| **resolved_human** | Escalated, human resolved the issue |
| **unresolved** | Escalated, issue not resolved |
| **unknown** | Session ended without clear outcome |

**Implicit Resolution:** If customer calls back about the same issue within 24 hours, original session marked as unresolved.

## 16.3 Analytics Flow

```
Conversation Ends
    |
    v
+-------------+     +-------------+     +-------------+
|   Store in   |---->|  Aggregate   |---->| Dashboard   |
|   PostgreSQL  |     |   (Queries)  |     |   (Read)    |
+-------------+     +-------------+     +-------------+
                          |
                          v
                   +-------------+
                   |   Export    |
                   |  (Parquet)  |
                   +-------------+
                          |
                          v
                   +-------------+
                   |   DuckDB     |
                   |  (Phase 2)  |
                   +-------------+
```

## 16.4 Logging Strategy

| Log Type | What | Retention |
|----------|------|-----------|
| **API Logs** | All requests, response times, org_id | 30 days |
| **Session Logs** | Session events, errors, tool calls | 90 days |
| **Audit Logs** | User actions, changes, admin operations | 1 year |
| **Error Logs** | Exceptions, failures, stack traces | 30 days |

---

# 17. Infrastructure & Deployment

## 17.1 Deployment Strategy

**Docker + Kamal (Self-hosted)**

```
+------------------+
|    Kamal         |
|  Deployment Tool  |
+------------------+
        |
        v
+------------------+
|    Docker         |
|   Containers       |
+------------------+
        |
        v
+------------------+
|  Application      |
|   Server(s)       |
+------------------+
```

## 17.2 Container Architecture

| Container | Purpose |
|-----------|---------|
| **api** | Go API server (REST + WebSocket) |
| **worker** | Background job processor (document processing, exports) |
| **redis** | Session state, caching, rate limiting |
| **postgres** | Primary database (with pgvector) |

## 17.3 Worker Container

The worker container handles all asynchronous background jobs. It is decoupled from the API server and runs as a separate process.

### Purpose

- Offload long-running tasks from the request path
- Ensure reliable job completion independent of client connections
- Enable horizontal scaling of background processing

### Job Queue

The job queue is backed by Redis. A dedicated Redis list acts as a FIFO queue:

```
Key: worker:jobs
Type: List
Operations: RPUSH (enqueue), BLPOP (dequeue)
```

Workers atomically claim jobs using `BLPOP` with timeout. Jobs include a type discriminator and payload:

```json
{
  "id": "job_uuid",
  "type": "audio_upload" | "document_processing" | "llm_judge" | "analytics_export" | "parquet_export" | "notification",
  "payload": { ... },
  "created_at": "ISO8601",
  "retry_count": 0
}
```

Failed jobs are re-enqueued with exponential backoff (max 3 retries).

### Worker Jobs

| Job Type | Trigger | Description |
|----------|---------|-------------|
| **audio_upload** | Session ends | Uploads session audio from temp disk to S3 |
| **document_processing** | Document uploaded | Validates, chunks, embeds, indexes into pgvector |
| **llm_judge** | Session ends | Runs post-session quality analysis |
| **analytics_export** | Scheduled (daily) | Generates Parquet exports for analytics |
| **parquet_export** | On demand | Exports requested report to S3 |
| **notification** | Various triggers | Sends async notifications (email, webhook) |

### Audio Upload Flow

```
Session Ends
    |
    v
API server enqueues: audio_upload job
    |
    v
Worker picks up job
    |
    v
Worker reads audio from temp disk: /tmp/sessions/{session_id}/audio.pcm
    |
    v
Worker uploads to S3: orgs/{org_id}/sessions/{session_id}/audio.pcm
    |
    v
On success: Update conversations.audio_url for all turns from this session
On failure: Re-enqueue with retry (max 3 attempts)
    |
    v
Temp file deleted after successful upload + 24h grace period
```

### LLM Judge Execution

The LLM Judge runs asynchronously after session ends:

```
Session Ends
    |
    v
API server enqueues: llm_judge job with session_id
    |
    v
Worker fetches full conversation transcript from PostgreSQL
    |
    v
Worker invokes LLM Judge (Gemini 2.5 Flash or Gemini 3 Flash, temperature 0)
    |
    v
Judge outputs:
  - Quality score (0.0-1.0)
  - Per-signal breakdown (confidence, sentiment, progress, kg_risk)
  - Knowledge gaps identified
  - Actionable feedback for org
    |
    v
Worker stores results in PostgreSQL (conversation_scores table)
    |
    v
Knowledge gaps queued as notification to org admin
```

**Output Storage Schema (conversation_scores table):**

```sql
CREATE TABLE conversation_scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    overall_score DECIMAL(3,2) NOT NULL CHECK (overall_score BETWEEN 0 AND 1),
    confidence_score DECIMAL(3,2) NOT NULL,
    sentiment_score DECIMAL(3,2) NOT NULL,
    progress_score DECIMAL(3,2) NOT NULL,
    kg_risk_score DECIMAL(3,2) NOT NULL,
    signal_breakdown JSONB DEFAULT '{}',
    suggested_action VARCHAR(50),
    knowledge_gaps JSONB DEFAULT '[]',
    agent_feedback TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_scores_session ON conversation_scores(session_id);
CREATE INDEX idx_scores_overall ON conversation_scores(overall_score);
```

---

## 17.4 Infrastructure Requirements

| Component | Requirement |
|-----------|-------------|
| **Redis** | Single instance for MVP; Sentinel/Cluster for scale |
| **PostgreSQL** | Single instance for MVP; Streaming replication for scale |
| **Application** | Stateless; horizontally scalable |
| **S3 Compatible** | External (MinIO/R2/Backblaze/S3) |

---

# 18. Non-Functional Requirements

## 18.1 Performance Targets

| Metric | Target |
|--------|--------|
| **Response Latency** | < 2 seconds for AI response |
| **Audio Latency** | < 500ms for voice (mic to speaker) |
| **Session Init** | < 2 seconds (with pre-warmed MCP) |
| **Concurrent Sessions** | 1,000 per project (Gemini limit) |
| **Session Recovery** | < 5 seconds reconnection |

## 18.2 Availability Targets

| Level | Target |
|-------|--------|
| **Platform** | 99.9% uptime |
| **Degraded Mode** | < 1 hour before escalation |

## 18.3 Security Requirements

| Requirement | Implementation |
|-------------|----------------|
| **Encryption at rest** | All data encrypted (database, Redis, S3) |
| **Encryption in transit** | TLS 1.3 |
| **Password hashing** | Argon2id (64MB, 3 iterations, 4 parallelism) |
| **MCP credentials** | AES-256-GCM encryption |
| **MFA** | Super Admin required, Org Admin recommended |
| **Audit logging** | All admin actions logged |

## 18.4 Scale Targets

| Metric | MVP Target | Scale Target |
|--------|-----------|--------------|
| **Concurrent calls** | 50 | 500 |
| **Orgs** | 10 | 100+ |
| **Documents per org** | 100 | 10,000 |
| **Sessions per day** | 1,000 | 50,000 |

---

# Appendix A: Glossary

| Term | Definition |
|------|------------|
| **API** | Application Programming Interface |
| **JWT** | JSON Web Token — stateless authentication token |
| **MCP** | Model Context Protocol — standard for AI tool integration |
| **RAG** | Retrieval-Augmented Generation — AI technique using external knowledge |
| **RPM** | Requests Per Minute |
| **S3** | Simple Storage Service — object storage |
| **TPM** | Tokens Per Minute |
| **UUID** | Universally Unique Identifier |
| **WebSocket** | Bidirectional real-time communication protocol |

---

# Appendix B: Open Decisions

| ID | Decision | Status |
|----|----------|--------|
| **OD-001** | Vector DB Provider | Locked: pgvector |
| **OD-002** | Deployment | Locked: Docker + Kamal |
| **OD-003** | S3 Provider | TBD |

---

| **OD-001** | Vector DB Provider | Locked: pgvector |
| **OD-002** | Deployment | Locked: Docker + Kamal |
| **OD-003** | S3 Provider | TBD |

---

*Document Status: Draft -- Pending Final Approval*
*Next Steps: Review with team, validate against PRD, confirm decisions*

---

# 19. Real-Time Conversation Scoring

## 19.1 Overview

Every turn during a live customer conversation, a lightweight **Classifier Agent** runs in parallel to the main Gemini Live session. It evaluates the conversation state and outputs a health score. This score drives live dashboard visibility and human takeover decisions.

**Model:** Gemini 2.5 Flash or Gemini 3 Flash at temperature 0
**Trigger:** Every conversation turn (customer message + AI response)
**Output:** Score 0.0-1.0 + per-signal breakdown + suggested action

---

## 19.2 Signal Taxonomy

### Confidence & Grounding (Weight: 40%)

Measures whether the AI is drawing from actual knowledge or hallucinating.

| Condition | Score |
|-----------|-------|
| RAG retrieval, chunk score > 0.75 | 1.0 |
| RAG retrieval, chunk score 0.50-0.75 | 0.7 |
| RAG retrieval, chunk score < 0.50 | 0.3 |
| No RAG context available | 0.2 |
| AI explicitly says "I don't know" / "I can't help" | 0.4 |
| MCP tool call succeeds | +0.15 bonus |
| MCP tool call fails | -0.20 penalty |

**Calculation:** Average across the last 5 turns. Cap at 1.0.

### Customer Sentiment & Frustration (Weight: 35%)

Measures whether the customer is satisfied or escalating.

| Condition | Score |
|-----------|-------|
| Explicit escalation request ("talk to a human," "frustrated") | 0.0 |
| High frustration markers ("doesn't work," "repeating," "why") | 0.1-0.3 |
| Mild skepticism ("okay but," "I'm not sure," "hmm") | 0.5 |
| Neutral / cooperative ("okay," "got it") | 0.7 |
| Satisfied / positive ("thanks," "perfect," "that helps") | 0.95 |

**Calculation:** Score the last customer message + the one before it. Average them.
**Override:** If explicit escalation request detected, sentiment score = 0.0 immediately, regardless of average.

### Progress Toward Resolution (Weight: 15%)

Measures whether the conversation is making headway.

| Turn Count | Base Score |
|------------|------------|
| Turn 1-3 | 1.0 |
| Turn 4-5 | 0.8 |
| Turn 6-8 | 0.6 |
| Turn 9-12 | 0.4 |
| Turn 13+ | 0.2 |

**Penalties:**

| Condition | Adjustment |
|-----------|------------|
| Customer asks the same question twice | -0.30 |
| AI contradicts itself | -0.20 |
| Customer says "I already told you" | -0.40 |

**Bonuses:**

| Condition | Adjustment |
|-----------|------------|
| Customer asks a follow-up building on prior answer | +0.10 |
| MCP tool call returns actionable result | +0.20 |

### Knowledge Gap Risk (Weight: 10%)

Measures likelihood the LLM Judge will flag this conversation post-session.

| Condition | Score |
|-----------|-------|
| All RAG chunks scored < 0.50 | 0.1 |
| MCP tool called but returned error | 0.2 |
| AI responds despite low confidence | 0.3 |
| RAG context found and used | 0.9 |
| Question matches known FAQ / common issue | 0.95 |

**Calculation:** Average of last 3 retrieved chunk relevance scores, capped at 1.0.

---

## 19.3 Overall Score Calculation

```
score = (0.40 x confidence) + (0.35 x sentiment) + (0.15 x progress) + (0.10 x kg_risk)
```

Result range: **0.0 - 1.0**

---

## 19.4 Threshold Tiers

Default thresholds (org-configurable):

| Range | Status | Action |
|-------|--------|--------|
| 0.70 - 1.00 | Green | Normal. Continue. |
| 0.50 - 0.69 | Yellow | Monitor. Flag on dashboard. Optional admin alert. |
| 0.30 - 0.49 | Orange | At risk. Dashboard notification. Suggest takeover. |
| 0.00 - 0.29 | Red | Critical. Auto-notify admin. Immediate takeover recommended. |

**Org-configurable overrides:**
- Red threshold: 0.2 (permissive) to 0.5 (aggressive)
- Auto-escalate on Red: toggle on/off per org
- Yellow/Orange thresholds: adjustable within platform bounds

---

## 19.5 Worked Example

**Context:** Turn 7, customer says "this still isn't working."

| Signal | Value | Weighted |
|--------|-------|---------|
| Confidence (RAG hit 0.8, no MCP) | 0.80 | 0.40 x 0.80 = 0.32 |
| Sentiment ("still isn't working") | 0.20 | 0.35 x 0.20 = 0.07 |
| Progress (Turn 7, no resolution) | 0.50 | 0.15 x 0.50 = 0.075 |
| KG Risk (chunks scored 0.6-0.7) | 0.65 | 0.10 x 0.65 = 0.065 |
| **Total** | | **0.53 -> Yellow** |

Dashboard flags the conversation. Admin can monitor or initiate takeover.

---

## 19.6 Human Takeover Flow

When a conversation is flagged (Orange or Red), or admin manually decides to intervene:

```
1. Admin sees flagged conversation on dashboard
   - Current score + which signal tanked
   - Last 3 turns preview
   - Suggested reason ("Customer frustrated," "AI out of knowledge," etc.)

2. Admin clicks "Take Over"

3. System:
   a. Pauses Gemini Live session gracefully
   b. Persists full conversation to PostgreSQL immediately
   c. Bundles classifier signals + context for the agent

4. Customer hears:
   "Thanks for your patience. A support specialist has reviewed our
    conversation and would like to help. One moment please..."

5. Human agent receives:
   - Full conversation transcript (read-only)
   - Per-signal score breakdown
   - Classifier's flagged reason
   - Session duration + turn count
   - Any knowledge gaps or failed tool calls

6. Agent resolves issue
   - Resolved -> session marked resolved_human
   - Further escalation needed -> marked escalated
```

---

## 19.7 Dashboard Real-Time View

Live conversations panel shows per-conversation:

| Field | Description |
|-------|-------------|
| Session ID | Unique identifier |
| Customer | Identifier if provided by org |
| Mode | Voice / Chat |
| Turn Count | Current turn number |
| Current Score | 0.0-1.0 with color indicator |
| Score Trend | Arrow: improving / stable / declining |
| Top Signal | Which signal is lowest right now |
| Status | Active / Flagged / Takeover Pending |

Admins can click any row to see the full live transcript + per-signal breakdown in real time.

---

## 19.8 Scoring Flow During Session

```
Customer Message
      |
      v
Gemini Live processes turn
      |
      v
Classifier Agent runs (parallel, async)
      |
      v
Score computed: 0.0 - 1.0
      |
      v
Score + breakdown stored in Redis (per session)
      |
      v
Dashboard polls / receives push update
      |
      v
If score < threshold:
   - Flag conversation
   - Notify admin (if auto-notify enabled)
   - Log suggested action
```

Scores are stored in Redis during the session for real-time access:

```
Key: session:{session_id}:score
Type: Hash
TTL: 30 minutes
Fields:
  - overall_score: decimal (0.0-1.0)
  - confidence: decimal
  - sentiment: decimal
  - progress: decimal
  - kg_risk: decimal
  - tier: green|yellow|orange|red
  - suggested_action: string
  - updated_at: timestamp
```

Post-session, final score is persisted to `conversation_scores` table by the worker.

---

# 20. System Prompt Generation

## 20.1 Overview

The system prompt is the base instruction injected into every Gemini Live session. It defines how the AI behaves, what tone it uses, and what constraints it follows.

System prompt generation has three modes, configurable per org:

| Mode | Description |
|------|-------------|
| **Auto-generate** | AI analyzes org's uploaded documents and generates a tailored prompt |
| **Auto+Edit** | AI generates a prompt; admin can review and customize before activation |
| **Custom** | Admin writes the system prompt from scratch |

---

## 20.2 System Prompt Generation Model

The system prompt is generated using **Gemini 2.5 Flash** (temperature 0.2 for slight creativity, deterministic output).

Inputs to the generator:
- All active documents from the org's knowledge base (text content only)
- Org name and industry (from org settings)
- Conversation context examples (if available)
- Custom instructions provided by admin (if mode is Auto+Edit or Custom)

Output: A structured system prompt stored in the org's configuration.

---

## 20.3 System Prompt Storage

The generated system prompt is stored in the `organizations.settings` JSONB field or in a dedicated table:

```sql
CREATE TABLE system_prompts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    mode VARCHAR(20) NOT NULL CHECK (mode IN ('auto', 'auto_edit', 'custom')),
    content TEXT NOT NULL,
    generator_version VARCHAR(20),
    is_active BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_prompts_org ON system_prompts(org_id);
CREATE INDEX idx_prompts_active ON system_prompts(org_id, is_active);
```

The active prompt is rebuilt on every session initialization from this storage.

---

## 20.4 System Prompt Modes

### Auto-generate

1. Admin enables auto-generate mode
2. System reads all active documents for the org
3. System invokes Gemini Flash to generate prompt
4. Prompt is activated immediately
5. Auto-regeneration triggered on: document upload, document update, admin request

### Auto+Edit

1. System generates prompt (same as auto-generate)
2. Admin reviews generated prompt in dashboard
3. Admin can edit custom instructions, tone, or constraints
4. Admin activates the prompt
5. Auto-regeneration updates the draft; admin must re-approve after edits

### Custom

1. Admin writes prompt from scratch in dashboard editor
2. System validates prompt (checks for unsafe content, required sections)
3. Admin activates custom prompt
4. No auto-regeneration unless admin requests it

---

## 20.5 System Prompt Structure

Generated and custom prompts follow this structure:

```
[Base Instruction]
You are a professional support agent for [Org Name].

[Behavior Rules]
- Be polite, concise, and helpful
- Use the customer's language
- Never make up information not in the knowledge base
- Escalate when unsure

[Tone & Style]
- Professional but friendly
- Short responses (voice support)
- Long-form responses (chat support)

[Context]
- Knowledge base: [dynamically injected from RAG]
- Available tools: [dynamically injected from MCP]

[Escalation]
- Escalate when: [org-configurable triggers]
```

---

## 20.6 System Prompt Regeneration Triggers

| Trigger | Action |
|---------|--------|
| New document uploaded | Mark prompt draft for regeneration |
| Document updated | Mark prompt draft for regeneration |
| Document deleted | Mark prompt draft for regeneration |
| Admin requests | Regenerate immediately |
| Weekly schedule | Auto-regenerate if docs changed |

Regeneration is asynchronous (worker job). Active sessions use the current prompt until session ends.
