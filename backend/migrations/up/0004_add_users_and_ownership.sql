-- Migration 0004: users table + task ownership (referential integrity)
--
-- Problem this fixes: the `tasks` table had no notion of "which user does
-- this task belong to", so there was no way to separate Yousaf's tasks from
-- Arham's or Aaqib's — everyone shared one flat list.
--
-- This migration:
--   1. Creates a `users` table.
--   2. Adds a nullable `user_id` column to `tasks` first (so it works on a
--      table that already has rows).
--   3. Backfills every existing task to a default user, so no row is left
--      without an owner.
--   4. Makes the column NOT NULL and adds the foreign key constraint, so
--      the database itself now enforces that every task belongs to a real
--      user (referential integrity), and deleting a user cascades to their
--      tasks instead of leaving orphans.

CREATE TABLE IF NOT EXISTS users (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name       VARCHAR(100)  NOT NULL,
  email      VARCHAR(150)  NOT NULL,
  created_at DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,

  UNIQUE KEY uq_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO users (id, name, email) VALUES
  (1, 'Yousaf', 'yousaf@example.com'),
  (2, 'Arham',  'arham@example.com'),
  (3, 'Aaqib',  'aaqib@example.com');

ALTER TABLE tasks
  ADD COLUMN user_id BIGINT UNSIGNED NULL AFTER id;

-- Every task that existed before this migration is assigned to user 1
-- (Yousaf) so no data is lost or left dangling.
UPDATE tasks SET user_id = 1 WHERE user_id IS NULL;

ALTER TABLE tasks
  MODIFY COLUMN user_id BIGINT UNSIGNED NOT NULL;

ALTER TABLE tasks
  ADD CONSTRAINT fk_tasks_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

CREATE INDEX idx_tasks_user_id ON tasks (user_id);
