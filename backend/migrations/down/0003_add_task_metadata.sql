-- Down-migration for 0003_add_task_metadata.sql
DROP INDEX idx_tasks_archived ON tasks;
DROP INDEX idx_tasks_category ON tasks;
DROP INDEX idx_tasks_priority ON tasks;

ALTER TABLE tasks
  DROP COLUMN archived,
  DROP COLUMN favorite,
  DROP COLUMN color,
  DROP COLUMN category,
  DROP COLUMN priority;
