-- Create function to get default user role
CREATE OR REPLACE FUNCTION get_default_role_id()
RETURNS UUID AS $$
BEGIN
    RETURN (SELECT id FROM roles WHERE name = 'User');
END;
$$ LANGUAGE plpgsql;

CREATE TABLE users (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role_id UUID NOT NULL DEFAULT get_default_role_id(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT NULL,
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_role_id ON users (role_id);
