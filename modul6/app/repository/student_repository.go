package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"modul6/app/model"
)

var (
	ErrNotFound  = errors.New("student not found")
	ErrDuplicate = errors.New("student already exists")
)

type StudentRepository interface {
	FindAll(ctx context.Context, params ListParams) ([]model.Student, int, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	Create(ctx context.Context, s model.Student) (model.Student, error)
	Update(ctx context.Context, s model.Student) (model.Student, error)
	Delete(ctx context.Context, id int) error
}

type ListParams struct {
	Page    int
	PerPage int
	Search  string
}

type studentRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentRepository{
		pool: pool,
	}
}

func (r *studentRepository) FindAll(
	ctx context.Context,
	params ListParams,
) ([]model.Student, int, error) {

	offset := (params.Page - 1) * params.PerPage

	search := strings.TrimSpace(params.Search)

	var (
		students []model.Student
		total    int
	)

	if search == "" {
		err := r.pool.QueryRow(
			ctx,
			`SELECT COUNT(*) FROM students`,
		).Scan(&total)

		if err != nil {
			return nil, 0, err
		}

		rows, err := r.pool.Query(
			ctx,
			`
			SELECT
				id,
				nim,
				name,
				email,
				grade,
				is_active,
				created_at,
				owner_id
			FROM students
			ORDER BY id
			LIMIT $1 OFFSET $2
			`,
			params.PerPage,
			offset,
		)

		if err != nil {
			return nil, 0, err
		}
		defer rows.Close()

		for rows.Next() {
			var s model.Student

			err := rows.Scan(
				&s.ID,
				&s.NIM,
				&s.Name,
				&s.Email,
				&s.Grade,
				&s.IsActive,
				&s.CreatedAt,
				&s.OwnerID,
			)

			if err != nil {
				return nil, 0, err
			}

			students = append(students, s)
		}

		if err := rows.Err(); err != nil {
			return nil, 0, err
		}

		return students, total, nil
	}

	searchPattern := "%" + search + "%"

	err := r.pool.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM students
		WHERE
			name ILIKE $1
			OR email ILIKE $1
			OR nim ILIKE $1
		`,
		searchPattern,
	).Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(
		ctx,
		`
		SELECT
			id,
			nim,
			name,
			email,
			grade,
			is_active,
			created_at,
			owner_id
		FROM students
		WHERE
			name ILIKE $1
			OR email ILIKE $1
			OR nim ILIKE $1
		ORDER BY id
		LIMIT $2 OFFSET $3
		`,
		searchPattern,
		params.PerPage,
		offset,
	)

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var s model.Student

		err := rows.Scan(
			&s.ID,
			&s.NIM,
			&s.Name,
			&s.Email,
			&s.Grade,
			&s.IsActive,
			&s.CreatedAt,
			&s.OwnerID,
		)

		if err != nil {
			return nil, 0, err
		}

		students = append(students, s)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return students, total, nil
}

func (r *studentRepository) FindByID(
	ctx context.Context,
	id int,
) (model.Student, error) {

	var s model.Student

	err := r.pool.QueryRow(
		ctx,
		`
		SELECT
			id,
			nim,
			name,
			email,
			grade,
			is_active,
			created_at,
			owner_id
		FROM students
		WHERE id = $1
		`,
		id,
	).Scan(
		&s.ID,
		&s.NIM,
		&s.Name,
		&s.Email,
		&s.Grade,
		&s.IsActive,
		&s.CreatedAt,
		&s.OwnerID,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Student{}, ErrNotFound
	}

	if err != nil {
		return model.Student{}, err
	}

	return s, nil
}

func (r *studentRepository) Create(
	ctx context.Context,
	s model.Student,
) (model.Student, error) {

	err := r.pool.QueryRow(
		ctx,
		`
		INSERT INTO students (
			nim,
			name,
			email,
			password,
			grade,
			is_active,
			owner_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING
			id,
			nim,
			name,
			email,
			password,
			grade,
			is_active,
			created_at,
			owner_id
		`,
		s.NIM,
		s.Name,
		s.Email,
		s.Password,
		s.Grade,
		s.IsActive,
		s.OwnerID,
	).Scan(
		&s.ID,
		&s.NIM,
		&s.Name,
		&s.Email,
		&s.Password,
		&s.Grade,
		&s.IsActive,
		&s.CreatedAt,
		&s.OwnerID,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return model.Student{}, ErrDuplicate
			}
		}

		return model.Student{}, err
	}

	return s, nil
}

func (r *studentRepository) Update(
	ctx context.Context,
	s model.Student,
) (model.Student, error) {

	err := r.pool.QueryRow(
		ctx,
		`
		UPDATE students
		SET
			nim = $1,
			name = $2,
			email = $3,
			grade = $4,
			is_active = $5
		WHERE id = $6
		RETURNING
			id,
			nim,
			name,
			email,
			grade,
			is_active,
			created_at,
			owner_id
		`,
		s.NIM,
		s.Name,
		s.Email,
		s.Grade,
		s.IsActive,
		s.ID,
	).Scan(
		&s.ID,
		&s.NIM,
		&s.Name,
		&s.Email,
		&s.Grade,
		&s.IsActive,
		&s.CreatedAt,
		&s.OwnerID,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Student{}, ErrNotFound
	}

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return model.Student{}, ErrDuplicate
			}
		}

		return model.Student{}, err
	}

	return s, nil
}

func (r *studentRepository) Delete(
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

func (r *studentRepository) String() string {
	return fmt.Sprintf("studentRepository")
}