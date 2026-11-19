CREATE TABLE link_analytics_events (
    id            BIGSERIAL PRIMARY KEY,
    link_id       BIGINT       NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    visitor_hash  VARCHAR(64)  NOT NULL,
    referrer_host VARCHAR(255) NOT NULL DEFAULT '',
    device_type   VARCHAR(16)  NOT NULL DEFAULT 'unknown',
    country_code  VARCHAR(2)  NOT NULL DEFAULT '',
    occurred_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX link_analytics_events_link_time_idx
    ON link_analytics_events (link_id, occurred_at);

CREATE INDEX link_analytics_events_link_visitor_idx
    ON link_analytics_events (link_id, visitor_hash, occurred_at);

CREATE INDEX link_analytics_events_time_idx
    ON link_analytics_events (occurred_at);
