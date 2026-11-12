-- +goose Up
CREATE SCHEMA ai_hub;

-- +goose Down
DROP SCHEMA IF EXISTS ai_hub CASCADE;
