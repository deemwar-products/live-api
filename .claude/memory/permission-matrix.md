---
name: permission-matrix
description: Granular permission matrix for org team members
metadata:
  type: project
---

# Permission Matrix - Org Team Members

## Decision
Org admins can assign granular permissions to their team members. This handles both small (1 user = all perms) and large (multiple users with scoped perms) organizations.

## Permission Categories

### Documents
- Upload
- Delete
- View
- Edit Metadata

### Knowledge Base
- Configure RAG
- Manage Chunks
- View Search Analytics

### AI / Voice
- Configure Voice Behavior
- Manage MCP Tools
- Configure Escalation Rules

### Team Management
- Invite Users
- Assign Permissions
- Remove Users
- View Team Activity

### Conversations
- View Transcripts
- Monitor Live Calls
- Take Over Escalated Calls

### Analytics
- View Dashboard
- Export Reports

### Settings
- Edit Org Profile
- Manage Webhooks

## Open Question
Should permissions be bundled into presets (e.g., "Agent" role bundling certain perms) OR keep fully granular (each permission toggle-able individually)?
