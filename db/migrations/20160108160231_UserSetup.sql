
-- +goose Up
create table users (
 id BIGSERIAL,
 user_name text,
 full_name text,
 password_hash text,
 is_enabled bool
);

-- +goose Down
DROP TABLE users;
