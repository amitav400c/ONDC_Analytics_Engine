-- Postgres schema for authentication
-- Seeded with a default admin account (password: admin123)

CREATE TABLE IF NOT EXISTS users (
    id         SERIAL PRIMARY KEY,
    email      VARCHAR(255) UNIQUE NOT NULL,
    password   VARCHAR(255) NOT NULL,  -- bcrypt hash
    name       VARCHAR(255) NOT NULL,
    role       VARCHAR(50) DEFAULT 'viewer',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Seed admin account
-- Password: admin123 (bcrypt hash)
INSERT INTO users (email, password, name, role) VALUES
    ('admin@ondc-analytics.dev', '$2a$10$42Tj2IO.dFMG7Qu5/tKodeZAd5Z6sBTzh9jEWfPTJBQRX3mv..T5u', 'Admin', 'admin')
ON CONFLICT (email) DO NOTHING;
