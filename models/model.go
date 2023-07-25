package model

type Book struct {
	ID            string `json:"id,omitempty" bson:"_id,omitempty"`
	Title         string `json:"title,omitempty" bson:"title,omitempty"`
	Author        string `json:"author,omitempty" bson:"author,omitempty"`
	NoOfChapters  int    `json:"no_of_chapters,omitempty" bson:"no_of_chapters,omitempty"`
	PublishedDate string `json:"published_date,omitempty" bson:"published_date,omitempty"`
}
