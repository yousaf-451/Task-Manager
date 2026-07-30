-- =====================================================================
-- Granet Technologies - Task Management Application
-- MySQL schema
--
-- Usage:
--   mysql -u root -p < schema.sql
--
-- This file is a one-shot convenience script for local setup (creates the
-- database, app user, tables, and seed data in one go). For anything beyond
-- the initial setup, treat migrations/up/*.sql as the source of truth for
-- schema changes going forward (see migrations/README.md).
-- =====================================================================

CREATE DATABASE IF NOT EXISTS task_manager
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

-- Dedicated application user (adjust password before running in production).
CREATE USER IF NOT EXISTS 'task_user'@'%' IDENTIFIED BY 'task_password';
GRANT ALL PRIVILEGES ON task_manager.* TO 'task_user'@'%';
FLUSH PRIVILEGES;

USE task_manager;

-- ---------------------------------------------------------------------
-- users table
--
-- Every task belongs to exactly one user. This is what lets the app tell
-- Yousaf's tasks apart from Arham's or Aaqib's: each row in `tasks` carries
-- a `user_id` that is a foreign key into this table (see below).
--
-- Authentication is real: POST /api/auth/signup stores a bcrypt
-- `password_hash` here, POST /api/auth/login verifies it and creates a row
-- in `sessions`, and every subsequent request is authenticated by the
-- session token cookie (see `sessions` table below and
-- internal/middleware/auth.go) rather than a self-reported header.
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
  id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name          VARCHAR(100)  NOT NULL,
  email         VARCHAR(150)  NOT NULL,
  password_hash VARCHAR(255)  NOT NULL DEFAULT '',
  created_at    DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,

  UNIQUE KEY uq_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------
-- sessions table
--
-- One row per logged-in browser session. `id` is the random session token
-- itself (also the value stored in the session cookie), so validating a
-- session is a primary-key lookup. Deleting a user cascades to their
-- sessions, same pattern as tasks below.
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS sessions (
  id         CHAR(64)        NOT NULL PRIMARY KEY,
  user_id    BIGINT UNSIGNED NOT NULL,
  created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expires_at DATETIME        NOT NULL,

  CONSTRAINT fk_sessions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,

  INDEX idx_sessions_user_id (user_id),
  INDEX idx_sessions_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------
-- tasks table
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS tasks (
  id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id      BIGINT UNSIGNED NOT NULL,
  title        VARCHAR(150)  NOT NULL,
  description  VARCHAR(2000) NOT NULL DEFAULT '',
  due_date     DATE          NOT NULL,
  status       ENUM('pending', 'in_progress', 'completed') NOT NULL DEFAULT 'pending',
  priority     ENUM('low', 'medium', 'high') NOT NULL DEFAULT 'medium',
  category     VARCHAR(60)   NOT NULL DEFAULT '',
  color        VARCHAR(9)    NOT NULL DEFAULT '#0e6b5c',
  favorite     BOOLEAN       NOT NULL DEFAULT FALSE,
  archived     BOOLEAN       NOT NULL DEFAULT FALSE,
  created_at   DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at   DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  -- Referential integrity: a task can never reference a user that doesn't
  -- exist, and deleting a user cleans up their tasks automatically instead
  -- of leaving orphan rows behind.
  CONSTRAINT fk_tasks_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,

  INDEX idx_tasks_user_id (user_id),
  INDEX idx_tasks_status (status),
  INDEX idx_tasks_due_date (due_date),
  INDEX idx_tasks_priority (priority),
  INDEX idx_tasks_category (category),
  INDEX idx_tasks_archived (archived)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------
-- Seed data (optional - safe to remove)
--
-- Three demo users so the "different users see different tasks" behavior
-- can be demonstrated immediately by switching the user in the UI.
-- ---------------------------------------------------------------------
INSERT INTO users (name, email) VALUES
  ('Yousaf', 'yousaf@example.com'),
  ('Arham',  'arham@example.com'),
  ('Aaqib',  'aaqib@example.com');

INSERT INTO tasks (user_id, title, description, due_date, status, priority, category, color, favorite) VALUES
  (1, 'Set up project repository', 'Initialize Git repo, add README and .gitignore', CURDATE(), 'completed', 'medium', 'Setup', '#0e6b5c', FALSE),
  (1, 'Design database schema', 'Model the tasks table with proper indexes', CURDATE() + INTERVAL 1 DAY, 'completed', 'high', 'Backend', '#0e6b5c', FALSE),
  (1, 'Build REST API', 'Implement CRUD endpoints in Go following layered architecture', CURDATE() + INTERVAL 3 DAY, 'in_progress', 'high', 'Backend', '#c98a1a', TRUE),
  (2, 'Build React frontend', 'Dashboard, task list, add/edit forms, filters', CURDATE() + INTERVAL 5 DAY, 'pending', 'medium', 'Frontend', '#2f6fed', FALSE),
  (3, 'Record demo video', 'Walk through the app and explain the architecture', CURDATE() + INTERVAL 7 DAY, 'pending', 'low', 'Docs', '#6b7280', FALSE);
