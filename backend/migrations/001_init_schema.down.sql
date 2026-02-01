-- Rollback initial database schema

DROP TABLE IF EXISTS migrations;
DROP TABLE IF EXISTS system_configs;
DROP TABLE IF EXISTS model_statistics;
DROP TABLE IF EXISTS billing_records;
DROP TABLE IF EXISTS payment_records;
DROP TABLE IF EXISTS user_api_keys;
DROP TABLE IF EXISTS models;
DROP TABLE IF EXISTS model_providers;
DROP TABLE IF EXISTS oauth_accounts;
DROP TABLE IF EXISTS oauth_providers;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS "uuid-ossp";
