package storage

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
	"uniqlo/app"
)

var DB *gorm.DB

func init() {
	var err error
	DB, err = gorm.Open(sqlite.Open("products.db"), &gorm.Config{})
	if err != nil {
		log.Println(err)
	}
	DB.AutoMigrate(&app.Product{})
}

func DeleteNotAvailableProducts() {
	DB.Delete(&app.Product{}, "last_seen < ?", time.Now().Add(time.Duration(-24)*time.Hour).Format("2006-01-02 15:04:05"))
}

func IsNewProduct(productID string) bool {
	var product app.Product
	DB.Where(&app.Product{Id: productID}).First(&product)
	return product.Id == ""
}

func CreateProduct(product app.Product) {
	DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"last_seen"}),
	}).Create(product)
}
