-- migrate:up

CREATE TYPE event_kind AS ENUM ('commit', 'account', 'identity');

CREATE TYPE commit_operation AS ENUM ('create', 'update', 'delete');

CREATE TABLE events (
    did TEXT,
    time_us TIMESTAMPTZ NOT NULL,
    kind event_kind NOT NULL,

    -- Commit fields
    rev TEXT,
	operation commit_operation,
	collection TEXT,
	rkey TEXT,
	record JSONB,
	cid TEXT,

    -- Account fields
	account_active BOOL,
	account_did TEXT,
	account_seq BIGINT,
	account_status TEXT,
	account_time TIMESTAMPTZ,

    -- Identity fields
	identity_did TEXT,
	identity_handle TEXT,
	identity_seq BIGINT,
	identity_time TIMESTAMPTZ,

    PRIMARY KEY(did, time_us)
);

CREATE INDEX idx_events_kind ON events (kind);
CREATE INDEX idx_events_did ON events (did);
CREATE INDEX idx_events_rev ON events (rev);
CREATE INDEX idx_events_operation ON events (operation);
CREATE INDEX idx_events_collection ON events ("collection");
CREATE INDEX idx_events_rkey ON events (rkey);
CREATE INDEX idx_events_cid ON events (cid);
CREATE INDEX idx_events_account_did ON events (account_did);
CREATE INDEX idx_events_account_seq ON events (account_seq);
CREATE INDEX idx_events_identity_did ON events (identity_did);
CREATE INDEX idx_events_identity_handle ON events (identity_handle);
CREATE INDEX idx_events_identity_seq ON events (identity_seq);


-- migrate:down
DROP TABLE events;
DROP TYPE event_kind;
DROP TYPE commit_operation;
