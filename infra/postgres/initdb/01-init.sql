-- Enable extensions
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pg_cron;

-- Verify extensions
SELECT * FROM pg_extension WHERE extname IN ('vector', 'pgcrypto', 'uuid-ossp', 'pg_cron');

-- Grant permissions to the connecting user. Uses current_user / current_database
-- so this works regardless of which POSTGRES_USER / POSTGRES_DB is configured.
DO $perms$
DECLARE
    u text := current_user;
    d text := current_database();
BEGIN
    EXECUTE format('GRANT ALL PRIVILEGES ON DATABASE %I TO %I', d, u);
    EXECUTE format('GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO %I', u);
    EXECUTE format('GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO %I', u);
END
$perms$;
