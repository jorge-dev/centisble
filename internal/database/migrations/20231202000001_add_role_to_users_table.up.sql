ALTER TABLE users
ADD COLUMN role VARCHAR(50) NOT NULL DEFAULT 'User' CHECK (role IN ('Admin', 'User', 'Guest', 'Moderator', 'Editor', 'Viewer', 'Manager'));
