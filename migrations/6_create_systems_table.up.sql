CREATE TABLE IF NOT EXISTS systems (
    id VARCHAR(191) PRIMARY KEY,
    name VARCHAR(255),
    url VARCHAR(255),
    secret_key VARCHAR(64),
    created_at BIGINT(19),
    updated_at BIGINT(19),
    INDEX systems_name(name),
    INDEX systems_secret_key(secret_key)
);
