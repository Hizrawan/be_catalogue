CREATE TABLE IF NOT EXISTS admin_access_tokens (
    id VARCHAR(191) PRIMARY KEY,
    admin_id VARCHAR(191),
    revoked_at DATETIME,
    expired_at DATETIME,
    created_at BIGINT(19),
    updated_at BIGINT(19),
    FOREIGN KEY (admin_id) REFERENCES admins(id)
);
