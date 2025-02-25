package models

type MovieFilters struct {
	SearchTerm string
	GenreId    string
	IsWatched  string
	Sort       string
}

type Movie struct {
	Id           int         `form:"id"`
	Title        string      `form:"title"`
	Description  string      `form:"description"`
	ReleaseYear  int         `form:"release_year"`
	Director     string      `form:"director"`
	Rating       int         `form:"rating"`
	IsWatched    bool        `form:"is_watched"`
	TrailerUrl   string      `form:"trailer_url"`
	PosterUrl    string      `form:"poster_url"`
	ViewsYouTube *int64      `form:"viewsYT"`
	VideoUrl     *string     `form:"video_url"`
	ViewsCount   *int        `form:"views_count"`
	Duration     *string     `form:"duration"`
	ScreenSrc    *string     `form:"screen_src"`
	Genres       []Genre     `form:"genres"`
	Category     []Category  `form:"categories"`
	Ages         []Age       `form:"ages"`
	AllSeries    []AllSeries `form:"allseries"`
}

type MovieUser struct {
	Id          int         `form:"id"`
	Title       string      `form:"title"`
	Description string      `form:"description"`
	ReleaseYear int         `form:"release_year"`
	Director    string      `form:"director"`
	TrailerUrl  string      `form:"trailer_url"`
	PosterUrl   string      `form:"poster_url"`
	Genres      []Genre     `form:"genres"`
	Category    []Category  `form:"categories"`
	Ages        []Age       `form:"ages"`
	AllSeries   []AllSeries `form:"allseries"`
}
