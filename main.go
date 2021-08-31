package main

import (
	"fmt"
	"uniqlo/app"
	"uniqlo/config"
	"uniqlo/storage"
	"uniqlo/telegram"
)

func main() {
	cfg := config.LoadConfigs()
	uniqlo := app.New()
	t := telegram.New(cfg.TelegramApiKey)
	storage.DeleteNotAvailableProducts()
	for _, category := range cfg.Categories {
		mainProducts := uniqlo.Scrape(category.URL)
		for _, mainProduct := range mainProducts {
			if !uniqlo.IsValidTitle(mainProduct.Title, cfg.NotValidTitles) {
				continue
			}
			products := uniqlo.ScrapeProductVariations(mainProduct.Id, mainProduct.Title)
			products = uniqlo.KeepValidOffersOnly(products, cfg)
			for _, product := range products {
				message := uniqlo.CreateMessage(product)
				if storage.IsNewProduct(product.Id) {
					err := t.SendMessage(category.Channel, message)
					if err != nil {
						fmt.Println(err)
					}
				}
				storage.CreateProduct(product)
			}
		}
	}
}
