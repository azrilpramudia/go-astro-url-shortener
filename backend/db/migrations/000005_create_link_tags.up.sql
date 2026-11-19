CREATE TABLE link_tags (
    id       BIGSERIAL PRIMARY KEY,
    link_id  BIGINT      NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    name     VARCHAR(32) NOT NULL,
    CONSTRAINT link_tags_link_id_name_key UNIQUE (link_id, name)
);

CREATE INDEX link_tags_name_link_id_idx ON link_tags (name, link_id);
