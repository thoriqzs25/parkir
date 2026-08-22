-- Drop all tables in reverse dependency order
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS incidents CASCADE;
DROP TABLE IF EXISTS transactions CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS shift_configs CASCADE;
DROP TABLE IF EXISTS rate_configs CASCADE;
DROP TABLE IF EXISTS gates CASCADE;
DROP TABLE IF EXISTS locations CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;

-- Drop trigger functions
DROP FUNCTION IF EXISTS check_rate_overlap();
DROP FUNCTION IF EXISTS check_shift_config_overlap();
DROP FUNCTION IF EXISTS update_updated_tz();

-- Drop extension
DROP EXTENSION IF EXISTS "pgcrypto";
