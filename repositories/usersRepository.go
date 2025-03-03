package repositories

import (
	"context"
	"goozinshe/models"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersRepository struct {
	db *pgxpool.Pool
}

func NewUsersRepository(conn *pgxpool.Pool) *UsersRepository {
	return &UsersRepository{db: conn}
}

func (r *UsersRepository) IsAdmin(c context.Context, id int) (bool, error) {
	var role string
	row := r.db.QueryRow(c, "select is_admin from users where id = $1", id)
	err := row.Scan(&role)
	if err != nil {
		return false, err
	}

	return role == "admin", nil
}

func (r *UsersRepository) FindById(c context.Context, id int) (models.User, error) {
	row := r.db.QueryRow(c, "select id, name, email, password, phonenumber, birthday, is_admin, poster_url from users where id = $1", id)

	var user models.User
	err := row.Scan(
		&user.Id,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.PhoneNumber,
		&user.Birthday,
		&user.IsAdmin,
		&user.PosterUrl)

	// PhoneNumber  int
	// Birthday     time.Tim

	return user, err
}

func (r *UsersRepository) FindByEmail(c context.Context, email string) (models.User, error) {
	row := r.db.QueryRow(c, "select id, name, email, password, phonenumber, birthday, is_admin, poster_url from users where email = $1", email)

	var user models.User
	err := row.Scan(
		&user.Id,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.PhoneNumber,
		&user.Birthday,
		&user.IsAdmin,
		&user.PosterUrl)
	if err == pgx.ErrNoRows {
		log.Println("Пользователь не найден")
		return models.User{}, nil
	} else if err != nil {
		log.Printf("Ошибка при поиске пользователя: %v", err)
		return models.User{}, err
	}
	return user, err
}

func (r *UsersRepository) FindAll(c context.Context) ([]models.User, error) {
	rows, err := r.db.Query(c, "select id, name, email, password, phonenumber, birthday, is_admin, poster_url from users order by id")
	if err != nil {
		return nil, err
	}

	users := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.Id,
			&user.Name,
			&user.Email,
			&user.Password,
			&user.PhoneNumber,
			&user.Birthday,
			&user.IsAdmin,
			&user.PosterUrl)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}
	if rows.Err() != nil {
		return nil, err
	}

	return users, nil
}

func (r *UsersRepository) Create(c context.Context, user models.User) (int, error) {
	var id int
	row := r.db.QueryRow(c, "insert into users(name, email, password, phonenumber, birthday, is_admin, poster_url) values($1, $2, $3, $4, $5, $6, $7) returning id",
		user.Name,
		user.Email,
		user.Password,
		user.PhoneNumber,
		user.Birthday,
		user.IsAdmin,
		user.PosterUrl)

	err := row.Scan(&id)
	if err == pgx.ErrNoRows {
		log.Println("Пользователь не найден")
		return 0, nil
	} else if err != nil {
		log.Printf("Ошибка при вставке пользователя: %v", err)
		return 0, err
	}

	return id, err
}

func (r *UsersRepository) UpdateUserProfiile(c context.Context, id int, updateUserProfiile models.User) error {
	_, err := r.db.Exec(c, "update users set name = $1, phonenumber = $2, birthday = $3,  poster_url = $4  where id = $5",
		updateUserProfiile.Name,
		updateUserProfiile.PhoneNumber,
		updateUserProfiile.Birthday,
		updateUserProfiile.PosterUrl, id)
	return err
}

func (r *UsersRepository) Update(c context.Context, id int, updateuser models.User) error {
	_, err := r.db.Exec(c, "update users set name = $1, email = $2, password = $3, phonenumber = $4, birthday = $5, is_admin = $6, poster_url = $7  where id = $8",
		updateuser.Name,
		updateuser.Email,
		updateuser.Password,
		updateuser.PhoneNumber,
		updateuser.Birthday,
		updateuser.IsAdmin,
		updateuser.PosterUrl, id)
	return err
}

func (r *UsersRepository) Delete(c context.Context, id int) error {
	_, err := r.db.Exec(c, "delete from users where id = $1", id)
	return err
}
