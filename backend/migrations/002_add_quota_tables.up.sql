-- Add quota management tables
-- Migration: 002_add_quota_tables

-- User quotas table - stores quota limits for each user
CREATE TABLE IF NOT EXISTS user_quotas (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    
    -- Daily limits
    daily_request_limit INTEGER NOT NULL DEFAULT 100,
    daily_token_limit INTEGER NOT NULL DEFAULT 100000,
    daily_cost_limit DECIMAL(10, 4) NOT NULL DEFAULT 10.00,
    
    -- Monthly limits
    monthly_request_limit INTEGER NOT NULL DEFAULT 3000,
    monthly_token_limit INTEGER NOT NULL DEFAULT 3000000,
    monthly_cost_limit DECIMAL(10, 4) NOT NULL DEFAULT 300.00,
    
    -- Model-specific limits (stored as JSON for flexibility)
    model_limits JSONB NOT NULL DEFAULT '{}',
    
    -- Rate limiting
    per_minute_rate_limit INTEGER NOT NULL DEFAULT 60,
    per_hour_rate_limit INTEGER NOT NULL DEFAULT 1000,
    
    -- Reset configuration
    reset_day INTEGER NOT NULL DEFAULT 1,
    timezone VARCHAR(100) NOT NULL DEFAULT 'UTC',
    
    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_quotas_user_id ON user_quotas(user_id);
CREATE INDEX idx_user_quotas_is_active ON user_quotas(is_active);

-- User daily usage table - tracks daily usage for quota enforcement
CREATE TABLE IF NOT EXISTS user_usage (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    
    -- Daily usage counts
    request_count INTEGER NOT NULL DEFAULT 0,
    token_count INTEGER NOT NULL DEFAULT 0,
    total_cost DECIMAL(10, 8) NOT NULL DEFAULT 0,
    
    -- Model-specific usage (stored as JSON for flexibility)
    model_usage JSONB NOT NULL DEFAULT '{}',
    
    -- Peak usage tracking
    peak_requests_per_minute INTEGER NOT NULL DEFAULT 0,
    peak_tokens_per_minute INTEGER NOT NULL DEFAULT 0,
    
    -- Status
    is_exceeded BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    UNIQUE(user_id, date)
);

CREATE INDEX idx_user_usage_user_id ON user_usage(user_id);
CREATE INDEX idx_user_usage_date ON user_usage(date);
CREATE INDEX idx_user_usage_is_exceeded ON user_usage(is_exceeded);
CREATE INDEX idx_user_usage_created_at ON user_usage(created_at);

-- Monthly usage summary table - for faster monthly limit checks
CREATE TABLE IF NOT EXISTS monthly_usage (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    year_month VARCHAR(7) NOT NULL, -- Format: "2025-01"
    
    -- Monthly usage counts
    request_count INTEGER NOT NULL DEFAULT 0,
    token_count INTEGER NOT NULL DEFAULT 0,
    total_cost DECIMAL(10, 4) NOT NULL DEFAULT 0,
    
    -- Model-specific usage (stored as JSON for flexibility)
    model_usage JSONB NOT NULL DEFAULT '{}',
    
    -- Status
    is_exceeded BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    UNIQUE(user_id, year_month)
);

CREATE INDEX idx_monthly_usage_user_id ON monthly_usage(user_id);
CREATE INDEX idx_monthly_usage_year_month ON monthly_usage(year_month);
CREATE INDEX idx_monthly_usage_is_exceeded ON monthly_usage(is_exceeded);

-- Insert default quotas for existing users
INSERT INTO user_quotas (user_id, created_at, updated_at)
SELECT 
    id, 
    NOW(), 
    NOW()
FROM users
ON CONFLICT (user_id) DO NOTHING;

-- Create initial daily usage records for existing users for today
INSERT INTO user_usage (user_id, date, created_at, updated_at)
SELECT 
    id, 
    CURRENT_DATE, 
    NOW(), 
    NOW()
FROM users
ON CONFLICT (user_id, date) DO NOTHING;

-- Create initial monthly usage records for existing users for current month
INSERT INTO monthly_usage (user_id, year_month, created_at, updated_at)
SELECT 
    id, 
    TO_CHAR(CURRENT_DATE, 'YYYY-MM'), 
    NOW(), 
    NOW()
FROM users
ON CONFLICT (user_id, year_month) DO NOTHING;