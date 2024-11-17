-- Remove the existing role column if it exists
ALTER TABLE users DROP COLUMN IF EXISTS role;

-- Add the new role_id column, initially allowing NULL
ALTER TABLE users ADD COLUMN role_id UUID;

-- Populate the new role_id column with the default role ID
DO $$
DECLARE
    default_role_id UUID;
BEGIN
    -- Ensure the default role exists and retrieve its ID
    SELECT id INTO default_role_id 
    FROM roles 
    WHERE name = 'User';

    IF default_role_id IS NULL THEN
        RAISE EXCEPTION 'Default role "User" not found in roles table';
    END IF;

    -- Update users table to set role_id for existing records
    UPDATE users
    SET role_id = default_role_id
    WHERE role_id IS NULL;
END $$;

-- Set the role_id column to NOT NULL
ALTER TABLE users ALTER COLUMN role_id SET NOT NULL;

-- Add a foreign key constraint between users.role_id and roles.id
ALTER TABLE users ADD CONSTRAINT fk_users_role
    FOREIGN KEY (role_id) 
    REFERENCES roles(id);

-- Create an index on the role_id column for better lookup performance
CREATE INDEX idx_users_role_id ON users(role_id);
