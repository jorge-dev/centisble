-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create schema if not exists
CREATE SCHEMA IF NOT EXISTS public;

-- Enable UUID-OSSP extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp" SCHEMA public;

-- Ensure search path is set correctly
SET search_path TO public;

