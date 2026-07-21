-- =====================================================================
-- Granet Technologies - Task Management Application
-- MySQL schema
--
-- Usage:
--   mysql -u root -p < schema.sql
--
-- This file is a one-shot convenience script for local setup (creates the
-- database, app user, table, and seed data in one go). For anything beyond
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
-- tasks table
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS tasks (
  id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
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

  INDEX idx_tasks_status (status),
  INDEX idx_tasks_due_date (due_date),
  INDEX idx_tasks_priority (priority),
  INDEX idx_tasks_category (category),
  INDEX idx_tasks_archived (archived)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------
-- Seed data (optional - safe to remove)
-- ---------------------------------------------------------------------
INSERT INTO tasks (title, description, due_date, status, priority, category, color, favorite) VALUES
  ('Set up project repository', 'Initialize Git repo, add README and .gitignore', CURDATE(), 'completed', 'medium', 'Setup', '#0e6b5c', FALSE),
  ('Design database schema', 'Model the tasks table with proper indexes', CURDATE() + INTERVAL 1 DAY, 'completed', 'high', 'Backend', '#0e6b5c', FALSE),
  ('Build REST API', 'Implement CRUD endpoints in Go following layered architecture', CURDATE() + INTERVAL 3 DAY, 'in_progress', 'high', 'Backend', '#c98a1a', TRUE),
  ('Build React frontend', 'Dashboard, task list, add/edit forms, filters', CURDATE() + INTERVAL 5 DAY, 'pending', 'medium', 'Frontend', '#2f6fed', FALSE),
  ('Record demo video', 'Walk through the app and explain the architecture', CURDATE() + INTERVAL 7 DAY, 'pending', 'low', 'Docs', '#6b7280', FALSE);
