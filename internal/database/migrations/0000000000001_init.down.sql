-- Disable extensions in reverse order
DROP EXTENSION IF EXISTS "citext";
DROP EXTENSION IF EXISTS "pgcrypto";
DROP EXTENSION IF EXISTS "uuid-ossp";

-- Reset timezone to default
SET timezone = 'GMT';

-- Drop schema (only if empty)
DROP SCHEMA IF EXISTS public CASCADE;
