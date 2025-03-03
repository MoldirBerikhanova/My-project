package repositories

import (
	"context"
	"fmt"
	"goozinshe/logger"
	"goozinshe/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GenresRepository struct {
	db *pgxpool.Pool
}

func NewGenresRepository(conn *pgxpool.Pool) *GenresRepository {
	return &GenresRepository{db: conn}
}

func (r *GenresRepository) IsAdmin(c context.Context, id int) (bool, error) {
	var role string
	row := r.db.QueryRow(c, "select is_admin from users where id = $1", id)
	err := row.Scan(&role)
	if err != nil {
		return false, err
	}

	return role == "admin", nil // проверяем, является ли роль "admin"
}

func (r *GenresRepository) FindById(c context.Context, id int) (models.Genre, error) {
	var genre models.Genre
	row := r.db.QueryRow(c, "select id, title, poster_url from genres where id = $1", id)
	err := row.Scan(&genre.Id, &genre.Title, &genre.PosterUrl)
	if err != nil {
		return models.Genre{}, err
	}

	return genre, nil
}

func (r *GenresRepository) FindAll(c context.Context) ([]models.Genre, error) {
	rows, err := r.db.Query(c, "select id, title, poster_url from genres")
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	genres := make([]models.Genre, 0)

	for rows.Next() {
		var genre models.Genre
		err = rows.Scan(&genre.Id, &genre.Title, &genre.PosterUrl)
		if err != nil {
			return nil, err
		}

		genres = append(genres, genre)
	}

	return genres, nil
}

func (r *GenresRepository) FindAllByIds(c context.Context, ids []int) ([]models.Genre, error) {
	rows, err := r.db.Query(c, "select id, title, poster_url from genres where id = any($1)", ids)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	genres := make([]models.Genre, 0)

	for rows.Next() {
		var genre models.Genre
		err = rows.Scan(&genre.Id, &genre.Title, &genre.PosterUrl)
		if err != nil {
			return nil, err
		}

		genres = append(genres, genre)
	}

	return genres, nil
}

func (r *GenresRepository) Create(c context.Context, genre models.Genre) (int, error) {
	var id int
	row := r.db.QueryRow(c, "insert into genres (title, poster_url) values ($1, $2) returning id", genre.Title, genre.PosterUrl)
	err := row.Scan(&id)
	if err != nil {
		return 0, nil
	}

	return id, nil
}

func (r *GenresRepository) Update(c context.Context, id int, updategenre models.Genre) error {
	_, err := r.db.Exec(c, "update genres set title = $1, poster_url = $2 where id = $3", updategenre.Title, updategenre.PosterUrl, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *GenresRepository) Delete(c context.Context, id int) error {
	l := logger.GetLogger()

	var genreTitle string
	row := r.db.QueryRow(c, "select title from genres where id = $1", id)
	err := row.Scan(&genreTitle)
	if err != nil {
		return err
	}

	l.Warn(fmt.Sprintf("Вы действительно хотите удалить %s жанр?", genreTitle))
	_, err = r.db.Exec(c, "delete from genres where id = $1", id)
	if err != nil {
		l.Error(err.Error())
		return err
	}

	l.Info(fmt.Sprintf("Жанр %s удален:", genreTitle))

	return nil
}

// var genre models.Genre
// err = rows.Scan(&genre.Id, &genre.Title, &genre.PosterUrl)
