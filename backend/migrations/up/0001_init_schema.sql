-- Migration 0001: initial schema
-- Applies to the `task_manager` database. When running via
-- docker-compose, MYSQL_DATABASE / MYSQL_USER already exist, so this
-- migration only creates the tasks table itself (no CREATE DATABASE/USER
-- here — see ../schema.sql for a self-contained local setup script).

CREATE TABLE IF NOT EXISTS tasks (
  id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  title        VARCHAR(150)  NOT NULL,
  description  VARCHAR(2000) NOT NULL DEFAULT '',
  due_date     DATE          NOT NULL,
  status       ENUM('pending', 'in_progress', 'completed') NOT NULL DEFAULT 'pending',
  created_at   DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at   DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  INDEX idx_tasks_status (status),
  INDEX idx_tasks_due_date (due_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
