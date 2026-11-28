-- +goose Up
CREATE TABLE ai_hub.ai_connection (
    id              uuid        PRIMARY KEY DEFAULT uuidv7(),
    domain_id       bigint      NOT NULL,
    name            text        NOT NULL,
    provider        text        NOT NULL,
    channel         text        NOT NULL DEFAULT 'consumer',
    mode            text        NOT NULL,
    auth_kind       text        NOT NULL,
    credentials     bytea,
    base_url        text,
    region          text,
    zero_retention  boolean     NOT NULL DEFAULT false,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    created_by      bigint,
    updated_by      bigint,
    CONSTRAINT ai_connection_name_uniq UNIQUE (domain_id, name),
    CONSTRAINT ai_connection_mode_chk CHECK (mode IN ('resold', 'byok', 'byom')),
    CONSTRAINT ai_connection_channel_chk CHECK (channel IN ('consumer', 'enterprise')),
    CONSTRAINT ai_connection_auth_chk
        CHECK (auth_kind IN ('api_key', 'gcp_service_account', 'azure_ad', 'none')),
    CONSTRAINT ai_connection_byom_url_chk
        CHECK (mode <> 'byom' OR base_url IS NOT NULL OR auth_kind = 'none'),
    CONSTRAINT ai_connection_enterprise_auth_chk
        CHECK (
            channel <> 'enterprise'
            OR (
                (provider = 'openai' AND auth_kind IN ('api_key', 'azure_ad'))
                OR (provider = 'gemini' AND auth_kind = 'gcp_service_account')
            )
        )
);

-- +goose StatementBegin
CREATE FUNCTION ai_hub.ai_connection_mode_immutable()
RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.mode IS DISTINCT FROM OLD.mode THEN
        RAISE EXCEPTION 'ai_connection.mode is immutable (connection %)', OLD.id
            USING ERRCODE = 'check_violation';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER ai_connection_mode_no_change
    BEFORE UPDATE OF mode ON ai_hub.ai_connection
    FOR EACH ROW EXECUTE FUNCTION ai_hub.ai_connection_mode_immutable();

-- +goose Down
DROP TRIGGER IF EXISTS ai_connection_mode_no_change ON ai_hub.ai_connection;
DROP FUNCTION IF EXISTS ai_hub.ai_connection_mode_immutable();
DROP TABLE IF EXISTS ai_hub.ai_connection;
