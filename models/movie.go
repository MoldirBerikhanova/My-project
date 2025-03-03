package models

type MovieFilters struct {
	SearchTerm string
	GenreId    string
	AgeId      string
	CategoryId string
	IsWatched  string
	Sort       string
}

type Movie struct {
	Id           int         `form:"id"`
	Title        string      `form:"title"`
	Description  string      `form:"description"`
	ReleaseYear  int         `form:"release_year"`
	Director     string      `form:"director"`
	Producer     *string     `form:"producer"`
	Rating       int         `form:"rating"`
	IsFavourite  bool        `form:"is_favourite"`
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
	AllSeries    []AllSeries `form:"allseries"` /// это сериалы без сезона
	Season       []Season    `form:"season"`    /// это с сезоном
}

type MovieUser struct {
	Id          int         `form:"id"`
	Title       string      `form:"title"`
	Description string      `form:"description"`
	ReleaseYear int         `form:"release_year"`
	Director    string      `form:"director"`
	Producer    *string     `form:"producer"`
	TrailerUrl  string      `form:"trailer_url"`
	PosterUrl   string      `form:"poster_url"`
	Genres      []Genre     `form:"genres"`
	Category    []Category  `form:"categories"`
	Ages        []Age       `form:"ages"`
	AllSeries   []AllSeries `form:"allseries"`
	Season      []Season    `form:"season"`
}

type MoviesAndSeasons struct {
	Movies  []Movie
	Seasons []Season
}
