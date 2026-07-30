-- Rollback for migration 0002. Removes only the known seed rows by title,
-- so it won't accidentally delete real user-created tasks with the same names.
DELETE FROM tasks WHERE title IN (
  'Set up project repository',
  'Design database schema',
  'Build REST API',
  'Build React frontend',
  'Record demo video'
);
