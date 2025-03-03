package repositories

import (
	"context"
	"fmt"
	"goozinshe/logger"
	"goozinshe/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(conn *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{db: conn}
}

func (r *CategoryRepository) IsAdmin(c context.Context, id int) (bool, error) {
	var role string
	row := r.db.QueryRow(c, "select role is_admin users where id = $1", id)
	err := row.Scan(&role)
	if err != nil {
		return false, err
	}

	return role == "admin", nil // проверяем, является ли роль "admin"
}

func (r *CategoryRepository) FindAllByIds(c context.Context, ids []int) ([]models.Category, error) {
	rows, err := r.db.Query(c, "select id, title, poster_url from categories where id = any($1)", ids)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	categories := make([]models.Category, 0)

	for rows.Next() {
		var category models.Category
		err = rows.Scan(&category.Id, &category.Title, &category.PosterUrl)
		if err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	return categories, nil
}

func (r *CategoryRepository) Create(c context.Context, category models.Category) (int, error) {
	var id int
	l := logger.GetLogger()
	// tx, err := r.db.Begin(c)
	l.Info(fmt.Sprintf("добавить категорию %s?", category.Title))
	row := r.db.QueryRow(c, "insert into categories (title, poster_url) values ($1, $2) returning id", category.Title, category.PosterUrl)
	err := row.Scan(&id)
	if err != nil {
		return 0, nil
	}

	return id, nil
}

func (r *CategoryRepository) FindById(c context.Context, id int) (models.Category, error) {
	var category models.Category
	row := r.db.QueryRow(c, "select id, title, poster_url from categories where id = $1", id)
	err := row.Scan(&category.Id, &category.Title, &category.PosterUrl)
	if err != nil {
		return models.Category{}, err
	}
	return category, nil
}

func (r *CategoryRepository) FindAll(c context.Context) ([]models.Category, error) {
	rows, err := r.db.Query(c, "select id, title, poster_url from categories")
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	categories := make([]models.Category, 0)
	for rows.Next() {
		var category models.Category
		err := rows.Scan(&category.Id, &category.Title, &category.PosterUrl)
		if err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	return categories, nil
}

func (r *CategoryRepository) Update(c context.Context, id int, updatedcategory models.Category) error {
	_, err := r.db.Exec(c, "update categories set title = $1, poster_url = $2 where id = $3", updatedcategory.Title, updatedcategory.PosterUrl, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *CategoryRepository) Delete(c context.Context, id int) error {
	l := logger.GetLogger()

	var categoryTitle string
	row := r.db.QueryRow(c, "select title from categories where id = $1", id)
	err := row.Scan(&categoryTitle)
	if err != nil {
		return err
	}

	l.Warn(fmt.Sprintf("Вы действительно хотите удалить %s категорию?", categoryTitle))
	_, err = r.db.Exec(c, "delete from categories where id = $1", id)
	if err != nil {
		return err
	}

	l.Info(fmt.Sprintf("категория %s удалена", categoryTitle))

	return nil
}
