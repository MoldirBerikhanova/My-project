package repositories

import (
	"context"
	"fmt"
	"goozinshe/logger"
	"goozinshe/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AgeRepository struct {
	db *pgxpool.Pool
}

func NewAgeRepository(conn *pgxpool.Pool) *AgeRepository {
	return &AgeRepository{db: conn}
}

func (r *AgeRepository) IsAdmin(c context.Context, id int) (bool, error) {
	var role string
	row := r.db.QueryRow(c, "select is_admin from users where id = $1", id)
	err := row.Scan(&role)
	if err != nil {
		return false, err
	}

	return role == "admin", nil // проверяем, является ли роль "admin"
}
func (r *AgeRepository) Create(c context.Context, age models.Age) (int, error) {
	l := logger.GetLogger()
	var id int
	l.Info(fmt.Sprintf("добавить возраст %s, перетащите картинку или загрузите %s", age.Age, age.PosterUrl))
	row := r.db.QueryRow(c, "insert into ages (age, poster_url) values($1, $2) returning id", age.Age, age.PosterUrl)
	err := row.Scan(&id)
	if err != nil {
		return 0, nil
	}
	l.Info(fmt.Sprintf("возраст %s создан", age.Age))
	return id, nil

}

func (r *AgeRepository) FindAll(c context.Context) ([]models.Age, error) {
	rows, err := r.db.Query(c, "select id, age, poster_url from ages")
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	ages := make([]models.Age, 0)

	for rows.Next() {
		var age models.Age
		err := rows.Scan(&age.Id, &age.Age, &age.PosterUrl)
		if err != nil {
			return nil, err
		}
		ages = append(ages, age)
	}
	return ages, nil
}

func (r *AgeRepository) FindById(c context.Context, id int) (models.Age, error) {
	var age models.Age
	row := r.db.QueryRow(c, "select id, age, poster_url from ages where id = $1", id)
	err := row.Scan(&age.Id, &age.Age, &age.PosterUrl)
	if err != nil {
		return models.Age{}, err
	}

	return age, nil
}

func (r *AgeRepository) FindAllByIds(c context.Context, ids []int) ([]models.Age, error) {
	rows, err := r.db.Query(c, "select id, age, poster_url from ages where id = any($1)", ids)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	ages := make([]models.Age, 0)

	for rows.Next() {
		var age models.Age
		err := rows.Scan(&age.Id, &age.Age, &age.PosterUrl)
		if err != nil {
			return nil, err
		}
		ages = append(ages, age)
	}
	return ages, nil
}

func (r *AgeRepository) Update(c context.Context, id int, updateage models.Age) error {
	_, err := r.db.Exec(c, "update ages set age = $1, poster_url = $2 where id = $3", updateage.Age, updateage.PosterUrl, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *AgeRepository) Delete(c context.Context, id int) error {
	l := logger.GetLogger()

	var ageTitle string
	row := r.db.QueryRow(c, "select age from ages where id = $1", id)
	err := row.Scan(&ageTitle)
	if err != nil {
		return err
	}

	l.Warn(fmt.Sprintf("Вы действительно хотите удалить %s возраст?", ageTitle))

	_, err = r.db.Exec(c, "delete from ages where id = $1", id)
	if err != nil {
		return err
	}

	l.Info(fmt.Sprintf("возраст %s удален", ageTitle))

	return nil
}
