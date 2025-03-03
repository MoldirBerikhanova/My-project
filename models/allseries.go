package models

type AllSeries struct {
	Id         *int    `form:"id"` // Указатель на int
	Series     *int    `form:"series"`
	Title      *string `form:"title"`
	TrailerUrl *string `form:"trailer_url"`
	Duration   *string `form:"duration"`
	PosterUrl  *string `form:"poster_url"`
}
