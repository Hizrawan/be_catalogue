package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"be20250107/utils/database"
)

type Specifications struct {
	ScreenSize  string `json:"screen_size"`
	FrontCamera string `json:"front_camera"`
	RearCamera  string `json:"rear_camera"`
	Ram         string `json:"ram"`
	Storage     string `json:"storage"`
	Battery     string `json:"battery"`
	OS          string `json:"os"`
	ProductCode string `json:"product_code"`
	Color       string `json:"color"`
}

type Catalogue struct {
	ID             int            `db:"id" json:"id"`
	Name           string         `db:"name" json:"name"`
	BrandID        int            `db:"brand_id" json:"brand_id"`
	BrandName      string         `db:"brand_name" json:"brand_name"`
	CategoryID     int            `db:"category_id" json:"category_id"`
	CategoryName   string         `db:"category_name" json:"category_name"`
	Specifications Specifications `db:"specifications" json:"specifications"`
	Price          float64        `db:"price" json:"price"`
	CreatedAt      time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt      *time.Time     `db:"deleted_at" json:"deleted_at"`
	PublishedAt    *time.Time     `db:"published_at" json:"published_at"`
	CreatedBy      string         `db:"created_by" json:"created_by"`
	UpdatedBy      string         `db:"updated_by" json:"updated_by"`
	DeletedBy      *string        `db:"deleted_by" json:"deleted_by"`
	Categories     []Category     `json:"categories"`
}

type Installment struct {
	ID           int       `db:"id" json:"id"`
	CatalogueID  int       `db:"catalogue_id" json:"catalogue_id"`
	ThreeMonths  float64   `db:"three_months" json:"three_months"`
	SixMonths    float64   `db:"six_months" json:"six_months"`
	TwelveMonths float64   `db:"twelve_months" json:"twelve_months"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type PriceHistory struct {
	ID          int       `db:"id" json:"id"`
	CatalogueID int       `db:"catalogue_id" json:"catalogue_id"`
	OldPrice    float64   `db:"old_price" json:"old_price"`
	NewPrice    float64   `db:"new_price" json:"new_price"`
	ChangedAt   time.Time `db:"changed_at" json:"changed_at"`
}

type Brand struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
type Category struct {
	ID          int        `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	Description string     `db:"description" json:"description"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at" json:"deleted_at"`
}

func (b *Brand) Bind(r *http.Request) error { return nil }

func (b *Brand) Insert(tx database.TxQueryer) error {
	query := `INSERT INTO brands (name) VALUES (:name);`
	_, err := tx.NamedExec(query, b)
	if err != nil {
		return fmt.Errorf("[Brand.Insert][NamedExec]%w", err)
	}
	var brandID int
	err = tx.QueryRow("SELECT LAST_INSERT_ID()").Scan(&brandID)
	if err != nil {
		return fmt.Errorf("[Brand.Insert][QueryRow]%w", err)
	}
	b.ID = brandID
	return nil
}

func (b *Brand) Update(tx database.TxQueryer) error {
	query := `UPDATE brands SET name = :name, updated_at = CURRENT_TIMESTAMP WHERE id = :id;`
	_, err := tx.NamedExec(query, b)
	if err != nil {
		return fmt.Errorf("[Brand.Update][NamedExec]%w", err)
	}
	return nil
}

func (b *Brand) Delete(tx database.TxQueryer) error {
	query := "DELETE FROM brands WHERE id = :id;"
	_, err := tx.NamedExec(query, b)
	if err != nil {
		return fmt.Errorf("[Brand.Delete][NamedExec]%w", err)
	}
	return nil
}

func GetBrands(db database.Queryer) ([]Brand, error) {
	brands := []Brand{}
	err := db.Select(&brands, "SELECT * FROM brands;")
	return brands, err
}

func GetBrand(db database.Queryer, id int) (Brand, error) {
	brand := Brand{}
	err := db.Get(&brand, "SELECT * FROM brands WHERE id=?;", id)
	return brand, err
}
func (c *Category) Bind(r *http.Request) error { return nil }
func (c *Category) Insert(tx database.TxQueryer) error {
	query := `INSERT INTO categories (name, description) VALUES (:name, :description);`
	_, err := tx.NamedExec(query, c)
	if err != nil {
		return fmt.Errorf("[Category.Insert][NamedExec]%w", err)
	}
	var categoryID int
	err = tx.QueryRow("SELECT LAST_INSERT_ID()").Scan(&categoryID)
	if err != nil {
		return fmt.Errorf("[Category.Insert][QueryRow]%w", err)
	}
	c.ID = categoryID
	return nil
}
func (c *Category) Update(tx database.TxQueryer) error {
	query := `UPDATE categories SET name = :name, description = :description, updated_at = CURRENT_TIMESTAMP WHERE id = :id;`
	_, err := tx.NamedExec(query, c)
	if err != nil {
		return fmt.Errorf("[Category.Update][NamedExec]%w", err)
	}
	return nil
}
func (c *Category) Delete(tx database.TxQueryer) error {
	query := "UPDATE categories SET deleted_at = CURRENT_TIMESTAMP WHERE id = :id;"
	_, err := tx.NamedExec(query, c)
	if err != nil {
		return fmt.Errorf("[Category.Delete][NamedExec]%w", err)
	}
	return nil
}
func GetCategories(db database.Queryer) ([]Category, error) {
	categories := []Category{}
	err := db.Select(&categories, "SELECT * FROM categories WHERE deleted_at IS NULL;")
	return categories, err
}
func (p *Catalogue) Bind(r *http.Request) error {
	return nil
}
func (p *Catalogue) Insert(tx database.TxQueryer) error {
	// Serialize Specifications
	specs, err := json.Marshal(p.Specifications)
	if err != nil {
		return fmt.Errorf("[Catalogue.Insert][Marshal Specifications]%w", err)
	}
	// Step 1: Insert into catalogues table and get the inserted ID
	query := ` 
		INSERT INTO catalogues (name, brand_id, category_id, specifications, price, created_by, updated_by) VALUES (:name, :brand_id, :category_id, :specifications, :price, :created_by, :updated_by); `
	result, err := tx.NamedExec(query, map[string]interface{}{"name": p.Name, "brand_id": p.BrandID, "category_id": p.CategoryID, "specifications": string(specs), "price": p.Price, "created_by": p.CreatedBy, "updated_by": p.UpdatedBy})
	if err != nil {
		return fmt.Errorf("[Catalogue.Insert][NamedExec]%w", err)
	}
	catalogueID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("[Catalogue.Insert][LastInsertId]%w", err)
	}
	p.ID = int(catalogueID)
	// Step 2: Insert into catalogues_categories table
	for _, category := range p.Categories {
		_, err := tx.Exec("INSERT INTO catalogues_categories (cata_id, cate_id) VALUES (?, ?)", p.ID, category.ID)
		if err != nil {
			return fmt.Errorf("[Catalogue.Insert][InsertCategory]%w", err)
		}
	}
	return nil
}

func (p *Catalogue) Update(tx database.TxQueryer) error {
	// Serialize Specifications
	specs, err := json.Marshal(p.Specifications)
	if err != nil {
		return fmt.Errorf("[Catalogue.Update][Marshal Specifications]%w", err)
	}

	// Get the old price
	var oldPrice float64
	err = tx.Get(&oldPrice, "SELECT price FROM catalogues WHERE id=?", p.ID)
	if err != nil {
		return fmt.Errorf("[Catalogue.Update][Get old price]%w", err)
	}

	// Insert price change into PriceHistory
	priceHistory := PriceHistory{
		CatalogueID: p.ID,
		OldPrice:    oldPrice,
		NewPrice:    p.Price,
		ChangedAt:   time.Now(),
	}
	err = priceHistory.Insert(tx)
	if err != nil {
		return fmt.Errorf("[Catalogue.Update][PriceHistory.Insert]%w", err)
	}

	// Update the Catalogue record
	query := `
    UPDATE catalogues SET name = :name, brand_id = :brand_id, specifications = :specifications, price = :price, published_at = :published_at, updated_at = CURRENT_TIMESTAMP
    WHERE id = :id;
  `
	_, err = tx.NamedExec(query, map[string]interface{}{
		"id":             p.ID,
		"name":           p.Name,
		"brand_id":       p.BrandID,
		"specifications": string(specs),
		"price":          p.Price,
		"published_at":   p.PublishedAt,
	})
	if err != nil {
		return fmt.Errorf("[Catalogue.Update][NamedExec]%w", err)
	}
	// Delete existing categories
	_, err = tx.Exec("DELETE FROM catalogues_categories WHERE cata_id = ?", p.ID)
	if err != nil {
		return fmt.Errorf("[Catalogue.Update][DeleteCategories]%w", err)
	}
	// Insert updated categories
	for _, tag := range p.Categories {
		_, err := tx.Exec("INSERT INTO catalogues_categories (cata_id, cate_id) VALUES (?, ?)", p.ID, tag.ID)
		if err != nil {
			return fmt.Errorf("[Catalogue.Update][InsertTag]%w", err)
		}
	}
	return nil
}

func (p *Catalogue) Delete(tx database.TxQueryer) error {
	query := "UPDATE catalogues SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?;"
	_, err := tx.Exec(query, p.ID)
	if err != nil {
		return fmt.Errorf("[Catalogue.Delete][Exec]%w", err)
	}
	return nil
}
func GetCatalogues(db database.Queryer, limit, offset int, sortBy, order, filterBy, filterValue string) ([]Catalogue, int, error) {
	Catalogues := []Catalogue{}
	baseQuery := `
       SELECT catalogues.id, catalogues.name, catalogues.brand_id, brands.name AS brand_name, catalogues.specifications, catalogues.price, catalogues.created_at, catalogues.updated_at, catalogues.deleted_at, catalogues.published_at 
       FROM catalogues 
       JOIN brands ON catalogues.brand_id = brands.id
       WHERE catalogues.deleted_at IS NULL
       `
	filterQuery := ""
	if filterBy != "" && filterValue != "" {
		filterQuery = fmt.Sprintf("AND %s LIKE '%%%s%%'", filterBy, filterValue)
	}
	sortQuery := ""
	if sortBy != "" && order != "" {
		sortQuery = fmt.Sprintf("ORDER BY %s %s", sortBy, order)
	}
	paginationQuery := ""
	if limit > 0 {
		paginationQuery = fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	}

	query := fmt.Sprintf("%s %s %s %s", baseQuery, filterQuery, sortQuery, paginationQuery)
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM catalogues JOIN brands ON catalogues.brand_id = brands.id WHERE catalogues.deleted_at IS NULL %s", filterQuery)

	var totalCount int
	err := db.Get(&totalCount, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("[GetCatalogues][Count]%w", err)
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, 0, fmt.Errorf("[GetCatalogues][Query]%w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var c Catalogue
		var specifications string
		err := rows.Scan(&c.ID, &c.Name, &c.BrandID, &c.BrandName, &specifications, &c.Price, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt, &c.PublishedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("[GetCatalogues][Scan]%w", err)
		}

		// Unmarshal the JSON stored in the 'specifications' field.
		err = json.Unmarshal([]byte(specifications), &c.Specifications)
		if err != nil {
			return nil, 0, fmt.Errorf("[GetCatalogues][Unmarshal Specifications]%w", err)
		}

		categories, err := GetCategoriesForCatalogue(db, c.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("[GetCatalogues][GetCategoriesForCatalogue]%w", err)
		}
		c.Categories = categories

		Catalogues = append(Catalogues, c)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("[GetCatalogues][Rows]%w", err)
	}

	return Catalogues, totalCount, nil
}
func GetCatalogue(db database.Queryer, id int) (Catalogue, error) {
	catalogue := Catalogue{}
	var specifications string
	query := `
    SELECT catalogues.id, catalogues.name, catalogues.brand_id, brands.name AS brand_name, catalogues.specifications, catalogues.price, catalogues.created_at, catalogues.updated_at, catalogues.deleted_at, catalogues.published_at 
    FROM catalogues 
    JOIN brands ON catalogues.brand_id = brands.id 
    WHERE catalogues.id = ? AND catalogues.deleted_at IS NULL;
    `
	err := db.QueryRow(query, id).Scan(&catalogue.ID, &catalogue.Name, &catalogue.BrandID, &catalogue.BrandName, &specifications, &catalogue.Price, &catalogue.CreatedAt, &catalogue.UpdatedAt, &catalogue.DeletedAt, &catalogue.PublishedAt)
	if err != nil {
		return Catalogue{}, fmt.Errorf("[GetCatalogue][Scan]%w", err)
	}

	// Unmarshal the JSON stored in the 'specifications' field.
	err = json.Unmarshal([]byte(specifications), &catalogue.Specifications)
	if err != nil {
		return Catalogue{}, fmt.Errorf("[GetCatalogue][Unmarshal Specifications]%w", err)
	}

	categories, err := GetCategoriesForCatalogue(db, catalogue.ID)
	if err != nil {
		return Catalogue{}, fmt.Errorf("[GetCatalogue][GetCategoriesForCatalogue]%w", err)
	}
	catalogue.Categories = categories

	return catalogue, nil
}

func GetCategoriesForCatalogue(db database.Queryer, CatalogueID int) ([]Category, error) {
	var categories []Category

	query := `
    SELECT t.id, t.name
    FROM categories t
    JOIN catalogues_categories pt ON t.id = pt.cate_id
    WHERE pt.cata_id = ?
  `
	err := db.Select(&categories, query, CatalogueID)
	if err != nil {
		return nil, fmt.Errorf("[GetCategoriesForCatalogue][Select]%w", err)
	}

	return categories, nil
}

func (i *Installment) Bind(r *http.Request) error {
	return nil
}

func (i *Installment) Insert(tx database.TxQueryer) error {
	query := `
    INSERT INTO installments (catalogue_id, three_months, six_months, twelve_months) 
    VALUES (:catalogue_id, :three_months, :six_months, :twelve_months);
  `
	_, err := tx.NamedExec(query, i)
	if err != nil {
		return fmt.Errorf("[Installment.Insert][NamedExec]%w", err)
	}
	return nil
}

func CalculateInstallments(price float64) Installment {
	return Installment{
		ThreeMonths:  price / 3,
		SixMonths:    price / 6,
		TwelveMonths: price / 12,
	}
}

func (ph *PriceHistory) Bind(r *http.Request) error {
	return nil
}

func (ph *PriceHistory) Insert(tx database.TxQueryer) error {
	query := `
    INSERT INTO price_history (catalogue_id, old_price, new_price, changed_at) 
    VALUES (:catalogue_id, :old_price, :new_price, :changed_at);
  `
	_, err := tx.NamedExec(query, ph)
	if err != nil {
		return fmt.Errorf("[PriceHistory.Insert][NamedExec]%w", err)
	}
	return nil
}
