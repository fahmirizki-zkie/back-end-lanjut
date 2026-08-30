package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"modul3/app/model"
)

var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data sudah ada")
)

type StudentRepository interface {
	FindAll(ctx context.Context) ([]model.Student, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	Create(ctx context.Context, s model.Student) (model.Student, error)
	Update(ctx context.Context, s model.Student) (model.Student, error)
	Delete(ctx context.Context, id int) error
}

type studentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentPostgresRepository{
		pool: pool,
	}
}

func (r *studentPostgresRepository) FindAll(
	ctx context.Context,
) ([]model.Student, error) {

	rows, err := r.pool.Query(ctx,
		`SELECT id, nim, name, grade, is_active, created_at
		 FROM students
		 ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	students := []model.Student{}

	for rows.Next() {
		var s model.Student

		if err := rows.Scan(
			&s.ID,
			&s.NIM,
			&s.Name,
			&s.Grade,
			&s.IsActive,
			&s.CreatedAt,
		); err != nil {
			return nil, err
		}

		students = append(students, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return students, nil
}

func (r *studentPostgresRepository) FindByID(
	ctx context.Context,
	id int,
) (model.Student, error) {

	var s model.Student

	err := r.pool.QueryRow(ctx,
		`SELECT id, nim, name, grade, is_active, created_at
		 FROM students
		 WHERE id = $1`,
		id,
	).Scan(
		&s.ID,
		&s.NIM,
		&s.Name,
		&s.Grade,
		&s.IsActive,
		&s.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}

		return model.Student{}, err
	}

	return s, nil
}

func (r *studentPostgresRepository) Create(
	ctx context.Context,
	s model.Student,
) (model.Student, error) {

	err := r.pool.QueryRow(ctx,
		`INSERT INTO students (nim, name, grade, is_active)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, nim, name, grade, is_active, created_at`,
		s.NIM,
		s.Name,
		s.Grade,
		s.IsActive,
	).Scan(
		&s.ID,
		&s.NIM,
		&s.Name,
		&s.Grade,
		&s.IsActive,
		&s.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.Student{}, ErrDuplicate
		}

		return model.Student{}, err
	}

	return s, nil
}

func (r *studentPostgresRepository) Update(
	ctx context.Context,
	s model.Student,
) (model.Student, error) {

	err := r.pool.QueryRow(ctx,
		`UPDATE students
		 SET nim = $1,
		     name = $2,
		     grade = $3,
		     is_active = $4
		 WHERE id = $5
		 RETURNING id, nim, name, grade, is_active, created_at`,
		s.NIM,
		s.Name,
		s.Grade,
		s.IsActive,
		s.ID,
	).Scan(
		&s.ID,
		&s.NIM,
		&s.Name,
		&s.Grade,
		&s.IsActive,
		&s.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.Student{}, ErrDuplicate
		}

		return model.Student{}, err
	}

	return s, nil
}

func (r *studentPostgresRepository) Delete(
	ctx context.Context,
	id int,
) error {

	tag, err := r.pool.Exec(
		ctx,
		`DELETE FROM students WHERE id = $1`,
		id,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}