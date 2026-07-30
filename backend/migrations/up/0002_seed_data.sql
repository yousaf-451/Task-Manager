-- Migration 0002: seed data for local development/demo purposes.
-- Safe to skip in production (see ../schema.sql notes).
INSERT INTO tasks (title, description, due_date, status) VALUES
  ('Set up project repository', 'Initialize Git repo, add README and .gitignore', CURDATE(), 'completed'),
  ('Design database schema', 'Model the tasks table with proper indexes', CURDATE() + INTERVAL 1 DAY, 'completed'),
  ('Build REST API', 'Implement CRUD endpoints in Go following layered architecture', CURDATE() + INTERVAL 3 DAY, 'in_progress'),
  ('Build React frontend', 'Dashboard, task list, add/edit forms, filters', CURDATE() + INTERVAL 5 DAY, 'pending'),
  ('Record demo video', 'Walk through the app and explain the architecture', CURDATE() + INTERVAL 7 DAY, 'pending');
