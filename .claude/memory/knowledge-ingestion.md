---
name: knowledge-ingestion
description: Knowledge base ingestion methods - current vs future scope
metadata:
  type: project
---

# Knowledge Base Ingestion

## In-scope (Phase 1)
1. **File Upload** - Upload PDFs, docs, FAQs directly
2. **Manual Entry** - Admins type/copy-paste content chunks
3. **API Integration** - Connect existing platforms (Notion, Confluence, Zendesk, etc.)

## Future Scope
- URL Scraping - auto-fetch content from URLs
- Businesses can bring their own MCP servers (with validation/sandboxing)
- Pre-built connectors: Notion, Google Calendar, Email, Generic API

## Technical Notes
- File processing: PDF parsing, doc extraction, chunking strategy (512 tokens, 64 overlap default)
- Chunk management: view, edit, delete individual chunks
- Embedding pipeline: extract → chunk → embed → store in vector DB (per-org namespace)
- API sync: webhook/Lambda triggers, or scheduled polling
