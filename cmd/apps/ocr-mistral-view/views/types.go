package views

// OCRData represents the structure of the Mistral OCR JSON data
type OCRData struct {
	Model     string     `json:"model"`
	Pages     []PageView `json:"pages"`
	UsageInfo struct {
		DocSizeBytes   int `json:"doc_size_bytes"`
		PagesProcessed int `json:"pages_processed"`
	} `json:"usage_info"`
}

// Page represents a single page in the OCR document
type PageView struct {
	Dimensions struct {
		DPI    int `json:"dpi"`
		Height int `json:"height"`
		Width  int `json:"width"`
	} `json:"dimensions"`
	Images   []Image `json:"images"`
	Index    int     `json:"index"`
	Markdown string  `json:"markdown"`
}

// Image represents an image within a page
type Image struct {
	BottomRightX int    `json:"bottom_right_x"`
	BottomRightY int    `json:"bottom_right_y"`
	ID           string `json:"id"`
	ImageBase64  string `json:"image_base64"`
	TopLeftX     int    `json:"top_left_x"`
	TopLeftY     int    `json:"top_left_y"`
}
