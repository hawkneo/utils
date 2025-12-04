-- +migrate Up
INSERT INTO test_schema (id, version, filename, hash, status) VALUES (3, 0, 'test file3', '0x01', 'applied');
INSERT INTO test_schema (id, version, filename, hash, status) VALUES (4, 0, 'test file4', '0x01', 'applied');

INSERT INTO test_schema (id, version, filename, hash, status) VALUES (5, 0, 'test file5', '0x01', 'applied');
INSERT INTO test_schema (id, version, filename, hash, status) VALUES (6, 0, 'test file6', '0x01', 'applied');

-- +migrate Up
INSERT INTO test_schema (id, version, filename, hash, status) VALUES (7, 0, 'test file7', '0x01', 'applied');
INSERT INTO test_schema (id, version, filename, hash, status) VALUES (8, 0, 'test file8', '0x01', 'applied');

-- +migrate Down
DELETE FROM test_schema WHERE id IN (3, 4, 5, 6);

-- +migrate Down
DELETE FROM test_schema WHERE id IN (7, 8);


-- +migrate Up
INSERT INTO test_schema (id, version, filename, hash, status) VALUES (10, 0, 'test file10', '0x01', 'applied');
INSERT INTO test_schema (id, version, filename, hash, status) VALUES (11, 0, 'test file11', '0x01', 'applied');
-- +migrate Down
DELETE FROM test_schema WHERE id IN (10, 11);
