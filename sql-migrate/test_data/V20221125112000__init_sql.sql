-- +migrate Up
CREATE TABLE test_schema
(
    id BIGSERIAL PRIMARY KEY NOT NULL ,
    version TEXT NOT NULL ,
    filename TEXT UNIQUE NOT NULL ,
    hash TEXT NOT NULL ,
    status TEXT NOT NULL ,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- +migrate Down
DROP TABLE test_schema;
