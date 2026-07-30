-- Migration 0005: real authentication (signup/login/logout + delete account)
--
-- Problem this fixes: "who is the current user" was previously just an
-- `X-User-Id` header the frontend set itself (see migration 0004 and the
-- User model's old doc comment) - anyone could claim to be any user id.
-- This migration adds real credentials and server-side sessions so a user
-- must prove who they are with a password, and every subsequent request is
-- authenticated by an unguessable session token instead of a self-reported
-- header.
--
-- 1. Adds `password_hash` to `users` (bcrypt hash, never the plaintext
--    password). Existing demo rows (Yousaf/Arham/Aaqib, seeded in 0004) get
--    an empty hash, which cannot match any real password - they simply
--    can't log in until/unless a password is set for them directly in the
--    database. That's acceptable for seed/demo data; every user created
--    from now on goes through POST /api/auth/signup and always has a real
--    password hash.
-- 2. Creates a `sessions` table: one row per logged-in browser session.
--    `id` is the random session token itself (also the cookie value), so
--    looking a session up is a primary-key lookup. Deleting a user cascades
--    to their sessions automatically, same pattern as tasks.

ALTER TABLE users
  ADD COLUMN password_hash VARCHAR(255) NOT NULL DEFAULT '' AFTER email;

CREATE TABLE IF NOT EXISTS sessions (
  id         CHAR(64)        NOT NULL PRIMARY KEY, -- random session token (hex-encoded)
  user_id    BIGINT UNSIGNED NOT NULL,
  created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expires_at DATETIME        NOT NULL,

  CONSTRAINT fk_sessions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,

  INDEX idx_sessions_user_id (user_id),
  INDEX idx_sessions_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
