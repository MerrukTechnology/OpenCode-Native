package page

// PageID is a unique identifier for a page
type PageID string

// PageChangeMsg is used to change the current page
type PageChangeMsg struct {
	ID PageID
}
