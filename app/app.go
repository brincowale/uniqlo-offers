package app

import (
	"github.com/go-resty/resty/v2"
	"github.com/gocolly/colly"
	"github.com/tidwall/gjson"
	"log"
	"regexp"
	"strconv"
	"time"
)

type Uniqlo struct {
	client *colly.Collector
}

type Product struct {
	id            string
	title         string
	color         string
	size          string
	stock         int64
	standardPrice float64
	salePrice     float64
	lastSeen      string
	url           string
	imageUrl      string
}

type MainProduct struct {
	id    string
	title string
}

func New() *Uniqlo {
	return &Uniqlo{
		client: colly.NewCollector(colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36")),
	}
}
func (u Uniqlo) Scrape(url string) []MainProduct {
	var mainProducts []MainProduct
	u.client.OnHTML("article.productTile", func(e *colly.HTMLElement) {
		p := MainProduct{
			id:    e.ChildAttr("a", "data-dlmasterid"),
			title: e.ChildAttr("a", "title"),
		}
		mainProducts = append(mainProducts, p)
	})
	u.client.Visit(url)
	return mainProducts
}

func (u Uniqlo) ScrapeProductVariations(mainProduct MainProduct) []Product {
	client := resty.New().SetTimeout(1*time.Minute).SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36")
	resp, err := client.R().Get("https://www.uniqlo.com/on/demandware.store/Sites-ES-Site/es_ES/Product-GetVariants?pid=" + mainProduct.id + "&Quantity=1")
	if err != nil {
		log.Println(err)
	}
	var products []Product
	result := gjson.Get(string(resp.Body()), "@this")
	for _, productVariation := range result.Map() {
		if !gjson.Get(productVariation.String(), "availability.inStock").Bool() {
			continue
		}
		var product Product
		product.id = gjson.Get(productVariation.String(), "id").String()
		product.title = mainProduct.title
		product.color = gjson.Get(productVariation.String(), "attributes.color").String()
		product.size = gjson.Get(productVariation.String(), "attributes.size").String()
		product.stock = gjson.Get(productVariation.String(), "availability.ats").Int()
		product.salePrice = gjson.Get(productVariation.String(), "pricing.sale").Float()
		product.standardPrice = gjson.Get(productVariation.String(), "pricing.standard").Float()
		product.lastSeen = time.Now().Format("2006-01-02 15:04:05")
		product.imageUrl = u.CreateProductImageURLVariation(product.id)
		product.url = u.CreateProductURLVariation(product.id)
		products = append(products, product)
	}
	return products

}

func (u Uniqlo) KeepValidOffersOnly(products []Product, minimumDiscount float64) []Product {
	var validProducts []Product
	for _, product := range products {
		percentage := 100 - (product.salePrice/product.standardPrice)*100
		if percentage >= minimumDiscount {
			validProducts = append(validProducts, product)
		}
	}
	return validProducts
}

func (u Uniqlo) CreateMessage(product Product) string {
	originalPrice := strconv.FormatFloat(product.standardPrice, 'f', -1, 64)
	discountedPrice := strconv.FormatFloat(product.salePrice, 'f', -1, 64)
	invisibleSpace := "&#8204;"
	returnLine := "\n"
	str :=
		product.title + " (" + product.color + ") [" + product.size + "]" + returnLine + "Antes: " + originalPrice +
			"€ - Ahora: " + discountedPrice + "€" + "<a href=\"" + product.imageUrl + "\">" + invisibleSpace + "</a>" +
			returnLine + "URL: " + product.url
	return str
}

func (u Uniqlo) CreateProductURLVariation(productID string) string {
	r, _ := regexp.Compile("^(\\d+)(COL\\d+)(\\w+\\d+)\\d{3}$")
	data := r.FindStringSubmatch(productID)
	return "https://www.uniqlo.com/es/es/" + string(data[1]) + ".html?dwvar_" + string(data[1]) + "_size=" + string(data[3]) + "%26dwvar_" + string(data[1]) + "_color=" + string(data[2])
}

func (u Uniqlo) CreateProductImageURLVariation(productID string) string {
	r, _ := regexp.Compile("^(\\d+)COL(\\d+)")
	data := r.FindStringSubmatch(productID)
	return "https://image.uniqlo.com/UQ/ST3/WesternCommon/imagesgoods/" + string(data[1]) + "/item/goods_" + string(data[2]) + "_" + string(data[1]) + ".jpg?width=500&impolicy=quality_60&imformat=chrome"
}
