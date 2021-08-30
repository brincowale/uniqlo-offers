package main

import (
	"fmt"
	"uniqlo/app"
	"uniqlo/config"
	"uniqlo/telegram"
)

func main() {
	cfg := config.LoadConfigs()
	fmt.Println(cfg.Categories)
	uniqlo := app.New()
	t := telegram.New(cfg.TelegramApiKey)
	//storage.DeleteNotAvailableProducts()
	for _, category := range cfg.Categories {
		mainProducts := uniqlo.Scrape(category.URL)
		for _, mainProduct := range mainProducts {
			products := uniqlo.ScrapeProductVariations(mainProduct)
			products = uniqlo.KeepValidOffersOnly(products, cfg.MinimumDiscount)
			for _, product := range products {
				message := uniqlo.CreateMessage(product)
				fmt.Println(message)
				fmt.Println(product)
				err := t.SendMessage(category.Channel, message)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}
