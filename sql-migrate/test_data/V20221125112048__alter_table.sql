-- +migrate Up
INSERT INTO test_schema (version, filename, hash, status) VALUES (0, 'test file', '0x00', 'applied');
