-- +goose Up
CREATE TABLE ai_hub.ai_model_catalog (
    provider      text NOT NULL,
    channel       text NOT NULL,
    kind          text NOT NULL,
    model_id      text NOT NULL,
    class         text NOT NULL,
    unit_type     text NOT NULL,
    capabilities  jsonb NOT NULL DEFAULT '{}'::jsonb,
    source        text NOT NULL DEFAULT 'seed',
    PRIMARY KEY (provider, channel, kind, model_id),
    CONSTRAINT catalog_source_chk CHECK (source IN ('seed', 'hub_sync', 'manual')),
    CONSTRAINT catalog_channel_chk CHECK (channel IN ('consumer', 'enterprise')),
    CONSTRAINT catalog_kind_chk
        CHECK (kind IN ('llm', 'tts', 'stt', 'embedding', 'realtime', 'reranker'))
);

-- +goose Down
DROP TABLE IF EXISTS ai_hub.ai_model_catalog;
