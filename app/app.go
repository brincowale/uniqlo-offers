package app

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/gocolly/colly"
	"github.com/tidwall/gjson"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
	"uniqlo/config"
)

type Uniqlo struct {
	client *colly.Collector
}

type Product struct {
	Id            string
	Title         string
	Color         string
	Size          string
	Stock         int64
	StandardPrice float64
	SalePrice     float64
	LastSeen      string
	Url           string
	ImageUrl      string
}

type MainProduct struct {
	Id    string
	Title string
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
			Id:    e.ChildAttr("a", "data-dlmasterid"),
			Title: e.ChildAttr("a", "title"),
		}
		mainProducts = append(mainProducts, p)
	})
	u.client.Visit(url)
	return mainProducts
}

func (u Uniqlo) ScrapeProductVariations(mainProductId string, mainProductTitle string) []Product {
	client := resty.New().SetTimeout(1*time.Minute).SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36")
	resp, err := client.R().Get("https://www.uniqlo.com/on/demandware.store/Sites-ES-Site/es_ES/Product-GetVariants?pid=" + mainProductId + "&Quantity=1")
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
		id := gjson.Get(productVariation.String(), "id").String()
		product.Title = mainProductTitle
		product.Color = gjson.Get(productVariation.String(), "attributes.color").String()
		product.Size = gjson.Get(productVariation.String(), "attributes.size").String()
		product.Stock = gjson.Get(productVariation.String(), "availability.ats").Int()
		product.SalePrice = gjson.Get(productVariation.String(), "pricing.sale").Float()
		product.StandardPrice = gjson.Get(productVariation.String(), "pricing.standard").Float()
		product.LastSeen = time.Now().Format("2006-01-02 15:04:05")
		product.ImageUrl = u.CreateProductImageURLVariation(id)
		product.Url = u.CreateProductURLVariation(id)
		product.Id = fmt.Sprintf("%s_%.2f", id, product.SalePrice)
		products = append(products, product)
	}
	return products

}

func (u Uniqlo) CreateMessage(product Product) string {
	originalPrice := strconv.FormatFloat(product.StandardPrice, 'f', -1, 64)
	discountedPrice := strconv.FormatFloat(product.SalePrice, 'f', -1, 64)
	invisibleSpace := "&#8204;"
	returnLine := "\n"
	str :=
		product.Title + " (" + product.Color + ") [" + product.Size + "]" + returnLine + "Antes: " + originalPrice +
			"€ - Ahora: " + discountedPrice + "€" + "<a href=\"" + product.ImageUrl + "\">" + invisibleSpace + "</a>" +
			returnLine + "Stock: " + strconv.FormatInt(product.Stock, 10) + returnLine + "URL: " + product.Url
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

func (u Uniqlo) KeepValidOffersOnly(products []Product, cfg *config.Config) []Product {
	var validProducts []Product
	for _, product := range products {
		if u.IsValidDiscount(product.SalePrice, product.StandardPrice, cfg.MinimumDiscount) &&
			u.IsValidTitle(product.Title, cfg.NotValidTitles) && u.IsValidSize(product.Size, cfg.NotValidSizes) {
			validProducts = append(validProducts, product)
		}
	}
	return validProducts
}

func (u Uniqlo) IsValidDiscount(salePrice float64, standardPrice float64, minimumDiscount float64) bool {
	percentage := 100 - (salePrice/standardPrice)*100
	return percentage >= minimumDiscount
}

func (u Uniqlo) IsValidSize(size string, notValidSizes []string) bool {
	for _, notValidSize := range notValidSizes {
		if strings.ToLower(notValidSize) == strings.ToLower(size) {
			return false
		}
	}
	return true
}

func (u Uniqlo) IsValidTitle(title string, notValidTitles []string) bool {
	title = strings.ToLower(title)
	for _, notValidTitle := range notValidTitles {
		if strings.Contains(title, strings.ToLower(notValidTitle)) {
			return false
		}
	}
	return true
}
