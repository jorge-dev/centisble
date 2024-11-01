-- Initialize common PostgreSQL extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";      -- For UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";       -- For cryptographic functions
CREATE EXTENSION IF NOT EXISTS "citext";         -- For case-insensitive text fields

-- Set timezone to UTC
SET timezone = 'UTC';
