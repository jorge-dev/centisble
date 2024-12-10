-- Revert timezone setting
RESET timezone;

-- Drop PostgreSQL extensions
DROP EXTENSION IF EXISTS "citext";
DROP EXTENSION IF EXISTS "pgcrypto";
DROP EXTENSION IF EXISTS "uuid-ossp";

-- Reset search path
RESET search_path;

-- Drop schema if exists
DROP SCHEMA IF EXISTS public CASCADE;
