package utils

type Product struct {
	ID				string		`gorm:"column:id"`
	Title			string		`gorm:"column:title"`
	ImageURL		string		`gorm:"column:image_url"`
	URL				string		`gorm:"column:url"`
	OriginalPrice	float64		`gorm:"column:original_price"`
	DiscountedPrice	float64		`gorm:"column:discounted_price"`
	LastSeen		string		`gorm:"column:last_seen"`
	Color			string		`gorm:"column:color"`
	Size			string		`gorm:"column:size"`
}

type Config struct {
	Database string
	ScrapeUrlList []string
	TelegramApiKey string
	TelegramChannelId string
}
