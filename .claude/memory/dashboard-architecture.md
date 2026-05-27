---
name: dashboard-architecture
description: Multi-tenant dashboard architecture - platform, business, customer
metadata:
  type: project
---

# Dashboard Architecture

## Three Interfaces

### 1. Platform Dashboard (Super Admin)
- **Who:** Core development team / platform provider
- **Access:** Full platform oversight - all orgs, super admin controls

### 2. Business Dashboard (Org Admin)
- **Who:** Companies who onboard our service
- **Access:** Manage their own:
  - Documents & Knowledge Base
  - Team Members & Permissions
  - AI/Voice Settings
  - Escalation Config (toggle ON/OFF)
  - View their analytics

### 3. Customer Web App
- **Who:** End-users seeking support
- **Access:** Interact with AI via chat/voice

## Key Principle
Business admins have control over their own customers - not platform-wide.

## Customer Web App Access
- **Subdomain:** `{org-slug}.yourplatform.com`
- **Embed/Integrate:** Companies can embed the chat widget into their own services/websites via snippet code or SDK

## Technical Implications
- Subdomain routing at API gateway level (tenant resolution)
- Dynamic branding per org (logo, colors, name)
- Embed snippet: JS widget that points to org subdomain or unique org ID