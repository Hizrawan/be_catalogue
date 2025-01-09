CREATE TABLE IF NOT EXISTS system_access_tokens (
    id VARCHAR(191) PRIMARY KEY,
    system_id VARCHAR(191),
    revoked_at DATETIME,
    expired_at DATETIME NOT NULL,
    created_at BIGINT(19),
    updated_at BIGINT(19),
    FOREIGN KEY (system_id) REFERENCES systems(id)
);