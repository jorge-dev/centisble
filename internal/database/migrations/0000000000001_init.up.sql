-- Create schema if not exists
CREATE SCHEMA IF NOT EXISTS public;

-- Enable UUID-OSSP extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp" SCHEMA public;

-- Ensure search path is set correctly
SET search_path TO public;

-- Initialize common PostgreSQL extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";      -- For UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";       -- For cryptographic functions
CREATE EXTENSION IF NOT EXISTS "citext";         -- For case-insensitive text fields

-- Set timezone to UTC
SET timezone = 'UTC';


