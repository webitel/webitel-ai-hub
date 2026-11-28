-- +goose Up
CREATE TABLE ai_hub.ai_profile (
    id                   uuid        PRIMARY KEY DEFAULT uuidv7(),
    domain_id            bigint      NOT NULL,
    connection_id        uuid        NOT NULL REFERENCES ai_hub.ai_connection(id) ON DELETE RESTRICT,
    name                 text        NOT NULL,
    kind                 text        NOT NULL,
    model_id             text        NOT NULL,
    model_class          text        NOT NULL,
    state                text        NOT NULL DEFAULT 'active',
    is_default           boolean     NOT NULL DEFAULT false,
    fallback_profile_id  uuid        REFERENCES ai_hub.ai_profile(id) ON DELETE SET NULL,
    config               jsonb       NOT NULL DEFAULT '{"schema_version":1,"params":{},"limits":{}}'::jsonb,
    embedding_dimensions integer,
    provider_options     jsonb       NOT NULL DEFAULT '{}'::jsonb,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    created_by           bigint,
    updated_by           bigint,
    CONSTRAINT ai_profile_name_uniq UNIQUE (domain_id, name),
    CONSTRAINT ai_profile_kind_chk
        CHECK (kind IN ('llm', 'tts', 'stt', 'embedding', 'realtime', 'reranker')),
    CONSTRAINT ai_profile_state_chk CHECK (state IN ('active', 'disabled', 'archived')),
    CONSTRAINT ai_profile_no_self_fallback_chk
        CHECK (fallback_profile_id IS DISTINCT FROM id),
    CONSTRAINT ai_profile_embedding_dims_chk
        CHECK (kind <> 'embedding' OR embedding_dimensions IS NOT NULL)
);

CREATE UNIQUE INDEX ai_profile_default_uniq
    ON ai_hub.ai_profile (domain_id, kind)
    WHERE is_default AND state = 'active';

-- +goose StatementBegin
CREATE FUNCTION ai_hub.ai_profile_fallback_depth_guard()
RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.fallback_profile_id IS NULL THEN
        RETURN NEW;
    END IF;

    IF EXISTS (
        SELECT 1 FROM ai_hub.ai_profile
        WHERE id = NEW.fallback_profile_id
          AND fallback_profile_id IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'fallback chain deeper than one hop (profile %)', NEW.id
            USING ERRCODE = 'check_violation';
    END IF;

    IF EXISTS (
        SELECT 1 FROM ai_hub.ai_profile WHERE fallback_profile_id = NEW.id
    ) THEN
        RAISE EXCEPTION 'profile % is a fallback target and cannot have its own fallback', NEW.id
            USING ERRCODE = 'check_violation';
    END IF;

    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER ai_profile_fallback_depth
    BEFORE INSERT OR UPDATE OF fallback_profile_id ON ai_hub.ai_profile
    FOR EACH ROW EXECUTE FUNCTION ai_hub.ai_profile_fallback_depth_guard();

-- +goose Down
DROP TRIGGER IF EXISTS ai_profile_fallback_depth ON ai_hub.ai_profile;
DROP FUNCTION IF EXISTS ai_hub.ai_profile_fallback_depth_guard();
DROP INDEX IF EXISTS ai_hub.ai_profile_default_uniq;
DROP TABLE IF EXISTS ai_hub.ai_profile;
