CREATE TABLE catalogues_categories (
    cata_id INT,
    cate_id INT,
    FOREIGN KEY (cata_id) REFERENCES catalogues(id),
    FOREIGN KEY (cate_id) REFERENCES categories(id),
    PRIMARY KEY (cata_id, cate_id)
);