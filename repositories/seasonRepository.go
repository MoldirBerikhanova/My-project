package repositories

import (
	"context"
	"fmt"
	"goozinshe/logger"
	"goozinshe/models"
	"log"
	"unsafe"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type SeasonRepository struct {
	db *pgxpool.Pool
}

func NewSeasonRepository(conn *pgxpool.Pool) *SeasonRepository {
	return &SeasonRepository{db: conn}
}

func (r *SeasonRepository) IsAdmin(c context.Context, id int) (bool, error) {
	var role string
	row := r.db.QueryRow(c, "select is_admin from users where id = $1", id)
	err := row.Scan(&role)
	if err != nil {
		return false, err
	}

	return role == "admin", nil
}

func (r *SeasonRepository) FindAllByIds(c context.Context, ids []int) ([]models.Season, error) {
	rows, err := r.db.Query(c, "select id, number,  title from season where id = any($1)", ids)
	defer rows.Close()
	if err != nil {
		l := logger.GetLogger()
		l.Error(err.Error())
		return nil, err
	}

	Season := make([]models.Season, 0)

	for rows.Next() {
		var season models.Season
		err = rows.Scan(
			&season.Id,
			&season.Number,
			&season.Title)
		if err != nil {
			l := logger.GetLogger()
			l.Error(err.Error())
			return nil, err
		}

		Season = append(Season, season)
	}

	return Season, nil
}

func (r *SeasonRepository) Create(c context.Context, season models.Season) (int, error) {
	var id int
	row := r.db.QueryRow(c, "insert into season (number, title) values($1, $2) returning id", season.Number, season.Title)
	err := row.Scan(&id)
	if err != nil {
		l := logger.GetLogger()
		l.Error(err.Error())
		return 0, err
	}
	return id, nil
}

func (r *SeasonRepository) FindById(c context.Context, seasonId int) (models.Season, error) {
	sql := `select 
			s.id,
			s.number,
			s.title,
			e.id, 
			e.series, 
			e.title,              
			e.trailer_url,
			e.duration,
			e.poster_url
		from season s
		JOIN seasons_allseries se ON se.season_id = s.id
		JOIN allseries e ON se.allserie_id = e.id
		where s.id = $1`
	l := logger.GetLogger()
	rows, err := r.db.Query(c, sql, seasonId)
	if err != nil {
		l.Error("Ошибка запроса к базе", zap.String("db_msg", err.Error()))
		return models.Season{}, err
	}
	var season models.Season
	found := false
	allseriesMap := make(map[int]*models.AllSeries)
	for rows.Next() {
		var s models.Season
		var e models.AllSeries
		err = rows.Scan(
			&s.Id,
			&s.Number,
			&s.Title,
			&e.Id,
			&e.Series,
			&e.Title,
			&e.TrailerUrl,
			&e.Duration,
			&e.PosterUrl,
		)
		if err != nil {
			l.Error(err.Error())
			return models.Season{}, err
		}
		if !found {
			season = s
			found = true
		}

		if e.Id != nil {
			if _, exists := allseriesMap[*e.Id]; !exists {
				allseriesMap[*e.Id] = &e
			}
		}
	}
	for _, v := range allseriesMap {
		season.AllSeries = append(season.AllSeries, *v)
	}
	return season, nil

}

func (r *SeasonRepository) FindAll(c context.Context) ([]models.Season, error) {
	l := logger.GetLogger()
	sql := `select 
			s.id,
			s.number,
			s.title,
			e.id, 
			e.series, 
			e.title,              
			e.trailer_url,
			e.duration,
			e.poster_url
		from season s
		JOIN seasons_allseries se ON se.season_id = s.id
		JOIN allseries e ON se.allserie_id = e.id`

	l.Info("Executing SQL query to fetch all seasons and series")
	rows, err := r.db.Query(c, sql)
	if err != nil {
		l.Error("Database query failed: " + err.Error())
		return nil, err
	}
	defer rows.Close()

	Season := make([]*models.Season, 0)
	seasonMap := make(map[int]*models.Season)
	allseriesMap := make(map[int]*models.AllSeries)
	l.Info("Processing query results")
	for rows.Next() {
		var s models.Season
		var e models.AllSeries

		err = rows.Scan(
			&s.Id,
			&s.Number,
			&s.Title,
			&e.Id,
			&e.Series,
			&e.Title,
			&e.TrailerUrl,
			&e.Duration,
			&e.PosterUrl,
		)
		if err != nil {
			log.Println("Error scanning row: " + err.Error())
			return nil, err
		}

		if _, exists := seasonMap[s.Id]; !exists {
			seasonMap[s.Id] = &s
			Season = append(Season, &s)
			l.Debug("Added series",
				zap.Int("SeasonID", s.Id),
				zap.Int("SeriesID", *e.Id),
				zap.String("SeasonMemoryAddress", fmt.Sprintf("%p", Season)))
		}

		if e.Id != nil {
			if _, exists := allseriesMap[*e.Id]; !exists {
				allseriesMap[*e.Id] = &e
				seasonMap[s.Id].AllSeries = append(seasonMap[s.Id].AllSeries, e)
				fmt.Printf("Added series: SeasonID=%d, SeriesID=%d, SeasonAddress=%p\n", s.Id, *e.Id, unsafe.Pointer(&Season))
			}
		}

	}

	err = rows.Err()
	if err != nil {
		l.Error("Rows iteration error: " + err.Error())
		return nil, err
	}

	log.Printf("Successfully fetched %d seasons", len(Season))

	concreteSeason := make([]models.Season, 0, len(Season))
	for _, v := range Season {
		concreteSeason = append(concreteSeason, *v)
	}

	return concreteSeason, nil
}

func (r *SeasonRepository) Update(c context.Context, seasonId int, season models.Season) error {
	_, err := r.db.Exec(c, `update season set 
							number = $1 ,
							title = $2
							where id = $3`,
		&season.Number,
		&season.Title,
		seasonId)
	if err != nil {
		l := logger.GetLogger()
		l.Error(err.Error())
		return err
	}

	return nil
}

// Id        int         `form:"id"`
// Number    int         `form:"number"`
// Title     string      `form:"title"`
// AllSeries []AllSeries `form:"allseries"`

//series, title, description, release_year, director, rating, trailer_url

func (r *SeasonRepository) Delete(c context.Context, seasonId int) error {
	l := logger.GetLogger()

	var seasonNumber string
	row := r.db.QueryRow(c, "select title from season where id = $1", seasonId)
	err := row.Scan(&seasonNumber)
	if err != nil {
		return err
	}

	l.Warn(fmt.Sprintf("Вы действительно хотите удалить %s сезон?", seasonNumber))

	_, err = r.db.Exec(c, "delete from Season where id = $1", seasonId)
	if err != nil {
		return err
	}

	l.Info(fmt.Sprintf("сезон %s удален", seasonNumber))

	return nil
}
