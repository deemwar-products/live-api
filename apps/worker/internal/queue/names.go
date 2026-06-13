package queue

// Stream names and consumer group names used across the worker and API.
const (
	StreamRag    = "jobs.rag"
	StreamWrites = "jobs.writes"

	GroupRag    = "workers.rag"
	GroupWrites = "workers.writes"
)

// Job type constants — values stored in jobs.type and carried in Redis messages.
const (
	TypeIngestDocument      = "INGEST_DOCUMENT"
	TypeUpdateJobStatus     = "UPDATE_JOB_STATUS"
	TypeUpdateDocumentStatus = "UPDATE_DOCUMENT_STATUS"
)
