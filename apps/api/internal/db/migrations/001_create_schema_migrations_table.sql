CREATE TABLE schema_migrations (
   version TEXT PRIMARY KEY,
   name TEXT NOT NULL,
   checksum TEXT NOT NULL,
   applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);