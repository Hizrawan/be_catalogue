CREATE TABLE IF NOT EXISTS roles(
    id VARCHAR(191) PRIMARY KEY,
    `name` VARCHAR(100) NOT NULL UNIQUE,
    created_at BIGINT(19),
    updated_at BIGINT(19)
);