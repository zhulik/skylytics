CREATE TABLE post_interactions (
   id SERIAL,

   cid VARCHAR NOT NULL, -- post id
   did VARCHAR NOT NULL, -- user id

   type VARCHAR(8) CHECK (type IN ('like', 'repost', 'reply')) NOT NULL,

   timestamp TIMESTAMPTZ NOT NULL
);

SELECT create_hypertable('post_interactions', 'timestamp');
SELECT add_retention_policy('post_interactions', INTERVAL '24 hours');

CREATE INDEX idx_post_interactions_type ON post_interactions (type);
CREATE INDEX idx_post_interactions_post ON post_interactions (cid);
