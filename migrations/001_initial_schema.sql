CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY,  
    username VARCHAR(255),  
    first_name VARCHAR(255),
    last_name VARCHAR(255),  
    is_bot BOOLEAN DEFAULT FALSE,
    language_code VARCHAR(10),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_last_seen ON users(last_seen);