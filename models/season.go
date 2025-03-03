package models

type Season struct {
	Id        int         `json:"id"`
	Number    int         `json:"number"`
	Title     string      `json:"title"`
	AllSeries []AllSeries `json:"allseries"`
}
