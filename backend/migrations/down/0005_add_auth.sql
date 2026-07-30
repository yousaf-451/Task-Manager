-- Rollback for migration 0005: drops sessions and the password_hash column.

DROP TABLE IF EXISTS sessions;

ALTER TABLE users
  DROP COLUMN password_hash;
