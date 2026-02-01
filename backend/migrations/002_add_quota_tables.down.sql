-- Remove quota management tables
-- Migration: 002_add_quota_tables

DROP TABLE IF EXISTS monthly_usage;
DROP TABLE IF EXISTS user_usage;
DROP TABLE IF EXISTS user_quotas;