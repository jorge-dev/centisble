
CREATE TABLE roles (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Insert default roles
INSERT INTO roles (name, description) VALUES
    ('Admin', 'Full system access and control'),
    ('User', 'Standard user access'),
    ('Guest', 'Limited access for guests'),
    ('Moderator', 'Content moderation capabilities'),
    ('Editor', 'Content editing access'),
    ('Viewer', 'Read-only access'),
    ('Manager', 'Department management access');
