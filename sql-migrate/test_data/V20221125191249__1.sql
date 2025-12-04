-- +migrate Up
INSERT INTO test_schema (id, version, filename, hash, status) VALUES (2, 0, 'test file2', '0x01', 'applied');

-- +migrate Down
DELETE FROM test_schema WHERE id = 2;
