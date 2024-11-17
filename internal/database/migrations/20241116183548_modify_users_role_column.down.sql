
-- Remove the foreign key and index
ALTER TABLE users DROP CONSTRAINT fk_users_role;
DROP INDEX IF EXISTS idx_users_role_id;

-- Remove role_id column
ALTER TABLE users DROP COLUMN role_id;

-- Add back the original role column
ALTER TABLE users 
ADD COLUMN role VARCHAR(50) NOT NULL DEFAULT 'User' 
CHECK (role IN ('Admin', 'User', 'Guest', 'Moderator', 'Editor', 'Viewer', 'Manager'));
