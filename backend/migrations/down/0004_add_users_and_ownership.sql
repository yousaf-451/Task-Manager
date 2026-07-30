-- Rollback for migration 0004.
ALTER TABLE tasks DROP FOREIGN KEY fk_tasks_user;
ALTER TABLE tasks DROP INDEX idx_tasks_user_id;
ALTER TABLE tasks DROP COLUMN user_id;
DROP TABLE IF EXISTS users;
