package storage

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"time"
)

var DB *gorm.DB

type Product struct {
	gorm.Model
	ProductID string `gorm:"size:50;index:idx_products_product_id,unique"`
	LastSeen  time.Time
}

func init() {
	var err error
	DB, err = gorm.Open(sqlite.Open("products.db"), &gorm.Config{})
	if err != nil {
		log.Println(err)
	}
	DB.AutoMigrate(&Product{})
}

func UpdateLastSeen(productID string) {
	DB.Where(&Product{ProductID: productID}).Update("last_seen", time.Now().Format("2006-01-02 15:04:05"))
}

func DeleteNotAvailableProducts() {
	DB.Delete(Product{}, "last_seen < ?", time.Now().Add(time.Duration(-24)*time.Hour).Format("2006-01-02 15:04:05"))
}
