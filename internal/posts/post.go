package posts

type Post struct {
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	UserID    uint   `json:"user_id"` // TODO: Make this a foreign key
	Upvotes   uint   `json:"upvotes"`
	Downvotes uint   `json:"downvotes"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
