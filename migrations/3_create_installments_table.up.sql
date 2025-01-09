CREATE TABLE IF NOT EXISTS installments (
    id INT AUTO_INCREMENT PRIMARY KEY,
    catalogue_id INT NOT NULL,
    installment_number INT NOT NULL,
    installment_amount DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255) NOT NULL,
    deleted_by VARCHAR(255) NULL,
    FOREIGN KEY (catalogue_id) REFERENCES catalogues(id)
);
