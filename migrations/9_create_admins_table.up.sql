CREATE TABLE IF NOT EXISTS admins (
    id VARCHAR(191) PRIMARY KEY,
    name VARCHAR(255),
    username VARCHAR(255),
    provider VARCHAR(255),
    role_id VARCHAR(191),
    provider_id VARCHAR(255),
    created_at BIGINT(19),
    updated_at BIGINT(19),
    deactivated_at DATETIME,
    FOREIGN KEY (role_id) REFERENCES roles(id),
    INDEX admins_username(username),
    INDEX admins_provider_id_provider(provider_id,provider)
);