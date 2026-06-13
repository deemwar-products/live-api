# live-api-poc local Postgres image — pgvector base + pg_cron.
#
# Build: docker compose build
# Run:   docker compose up -d

FROM pgvector/pgvector:pg16

ARG PG_VERSION=16

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
        postgresql-${PG_VERSION}-cron \
 && rm -rf /var/lib/apt/lists/*
