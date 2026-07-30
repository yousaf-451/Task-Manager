-- Migration 0003: task metadata (priority, category, color, favorite, archived)
--
-- Backward compatible: every new column has a default, so rows written by
-- the pre-migration API (and any client that hasn't been updated yet)
-- remain valid. No existing column is renamed or dropped.

ALTER TABLE tasks
  ADD COLUMN priority ENUM('low', 'medium', 'high') NOT NULL DEFAULT 'medium' AFTER status,
  ADD COLUMN category  VARCHAR(60)  NOT NULL DEFAULT '' AFTER priority,
  ADD COLUMN color     VARCHAR(9)   NOT NULL DEFAULT '#0e6b5c' AFTER category,
  ADD COLUMN favorite  BOOLEAN      NOT NULL DEFAULT FALSE AFTER color,
  ADD COLUMN archived  BOOLEAN      NOT NULL DEFAULT FALSE AFTER favorite;

CREATE INDEX idx_tasks_priority ON tasks (priority);
CREATE INDEX idx_tasks_category ON tasks (category);
CREATE INDEX idx_tasks_archived ON tasks (archived);
