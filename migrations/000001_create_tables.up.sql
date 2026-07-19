CREATE TABLE summaries (
    id UUID PRIMARY KEY,
    external_summary_id TEXT NOT NULL,
    host TEXT NOT NULL,
    port INTEGER NOT NULL,
    user_name TEXT NOT NULL,
    db_name TEXT NOT NULL,
    synced_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE schemas (
    id UUID PRIMARY KEY,
    summary_id UUID NOT NULL,
    name TEXT NOT NULL,

    CONSTRAINT fk_summary
        FOREIGN KEY (summary_id)
        REFERENCES summaries(id)
        ON DELETE CASCADE
);

CREATE TABLE tables (
    id UUID PRIMARY KEY,
    schema_id UUID NOT NULL,
    name TEXT NOT NULL,
    row_count BIGINT NOT NULL,
    size_mb DOUBLE PRECISION NOT NULL,

    CONSTRAINT fk_schema
        FOREIGN KEY (schema_id)
        REFERENCES schemas(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_schemas_summary_id
ON schemas(summary_id);

CREATE INDEX idx_tables_schema_id
ON tables(schema_id);