ALTER TABLE users
ADD COLUMN IF NOT EXISTS system_role VARCHAR(32) NOT NULL DEFAULT 'user';

UPDATE users
SET system_role = 'user'
WHERE system_role IS NULL OR system_role = '';
