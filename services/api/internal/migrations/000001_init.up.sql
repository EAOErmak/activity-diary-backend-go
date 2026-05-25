CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  username VARCHAR(255) NOT NULL UNIQUE,
  full_name VARCHAR(255) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  role VARCHAR(32) NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash VARCHAR(255) NOT NULL,
  revoked BOOLEAN NOT NULL DEFAULT FALSE,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS diary_entries (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  when_started TIMESTAMPTZ NULL,
  when_ended TIMESTAMPTZ NULL,
  duration INTEGER NULL,
  mood INTEGER NULL,
  description TEXT NULL,
  status VARCHAR(32) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tags (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL UNIQUE,
  status VARCHAR(32) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS diary_entry_tags (
  diary_entry_id BIGINT NOT NULL REFERENCES diary_entries(id) ON DELETE CASCADE,
  tag_id BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY (diary_entry_id, tag_id)
);

CREATE TABLE IF NOT EXISTS dictionary_items (
  id BIGSERIAL PRIMARY KEY,
  type VARCHAR(32) NOT NULL,
  label VARCHAR(255) NOT NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  allowed_role VARCHAR(32) NULL,
  parent_id BIGINT NULL REFERENCES dictionary_items(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dictionary_items_type_label ON dictionary_items(type, label);

CREATE TABLE IF NOT EXISTS diary_entry_metrics (
  id BIGSERIAL PRIMARY KEY,
  diary_entry_id BIGINT NOT NULL REFERENCES diary_entries(id) ON DELETE CASCADE,
  metric_type_id BIGINT NOT NULL REFERENCES dictionary_items(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS diary_entry_metric_values (
  id BIGSERIAL PRIMARY KEY,
  diary_entry_metric_id BIGINT NOT NULL REFERENCES diary_entry_metrics(id) ON DELETE CASCADE,
  unit_id BIGINT NOT NULL REFERENCES dictionary_items(id),
  value NUMERIC NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS metric_name_unit_links (
  metric_name_id BIGINT NOT NULL REFERENCES dictionary_items(id) ON DELETE CASCADE,
  metric_unit_id BIGINT NOT NULL REFERENCES dictionary_items(id) ON DELETE CASCADE,
  PRIMARY KEY (metric_name_id, metric_unit_id)
);

CREATE TABLE IF NOT EXISTS tag_metric_links (
  tag_id BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  metric_name_id BIGINT NOT NULL REFERENCES dictionary_items(id) ON DELETE CASCADE,
  PRIMARY KEY (tag_id, metric_name_id)
);

CREATE TABLE IF NOT EXISTS tag_chart_type_links (
  tag_id BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  chart_type VARCHAR(64) NOT NULL,
  PRIMARY KEY (tag_id, chart_type)
);

CREATE TABLE IF NOT EXISTS general_foods (
  id BIGSERIAL PRIMARY KEY,
  dictionary_item_id BIGINT NOT NULL UNIQUE REFERENCES dictionary_items(id) ON DELETE CASCADE,
  protein NUMERIC NOT NULL,
  fat NUMERIC NOT NULL,
  carbs NUMERIC NOT NULL,
  callories NUMERIC NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
