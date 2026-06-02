---
name: analytics-requirements
description: Analytics and reporting requirements for business dashboard
metadata:
  type: project
---

# Analytics & Reporting

## Metrics to Track

| Category | Metrics |
|----------|----------|
| **Volume** | Total calls/chats, per period (daily/weekly/monthly) |
| **Escalation** | % escalated, escalation reasons |
| **AI Performance** | Answer success rate, topics AI struggles with |
| **Knowledge Gaps** | Flagged content that couldn't answer |
| **User Satisfaction** | Feedback, ratings |

## Additional Insights
- Peak hours (help businesses staff accordingly)
- Failed queries (questions AI couldn't answer)
- Topics breakdown (what customers ask about most)

## Data Storage
- PostgreSQL for structured data
- DuckDB + S3 Parquet for analytics at scale (millions of rows)

## Retention
- TBD based on plan tier (basic/pro/enterprise)