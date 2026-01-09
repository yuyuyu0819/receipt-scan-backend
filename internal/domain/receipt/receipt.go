package receipt

// Receipt represents a stored receipt linked to a user.
type Receipt struct {
	ID       int64  `json:"id"`
	UserName string `json:"userName"`
	Store    string `json:"store"`
	Date     string `json:"date"`
	Total    int    `json:"total"`
}
