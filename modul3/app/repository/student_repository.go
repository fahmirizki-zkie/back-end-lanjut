package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"modul3/app/model"
)

var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data sudah ada")
)

type StudentRepository interface {
	FindAll(ctx context.Context, params ListParams) ([]model.Student, int, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	Create(ctx context.Context, s model.Student) (model.Student, error)
	Update(ctx context.Context, s model.Student) (model.Student, error)
	Delete(ctx context.Context, id int) error
}

type ListParams struct {
	Search   string
	IsActive *bool
	Sort     string
	Order    string
	Limit    int
	Offset   int
}

type studentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentPostgresRepository{
		pool: pool,
	}
}

// FindAll mengambil data student dengan:
// search, filter, sorting, pagination, dan total count
func (r *studentPostgresRepository) FindAll(
	ctx context.Context,
	params ListParams,
) ([]model.Student, int, error) {

	// Whitelist kolom untuk ORDER BY
	allowedSort := map[string]string{
		"id":         "id",
		"name":       "name",
		"nim":        "nim",
		"grade":      "grade",
		"is_active":  "is_active",
		"created_at": "created_at",
	}

	sortColumn, ok := allowedSort[params.Sort]
	if !ok {
		sortColumn = "id"
	}

	order := "ASC"

	if strings.ToLower(params.Order) == "desc" {
		order = "DESC"
	}

	// WHERE
	where := []string{}
	args := []any{}

	// Search nama menggunakan ILIKE
	if params.Search != "" {
		args = append(args, "%"+params.Search+"%")

		where = append(
			where,
			fmt.Sprintf("name ILIKE $%d", len(args)),
		)
	}

	// Filter is_active
	if params.IsActive != nil {
		args = append(args, *params.IsActive)

		where = append(
			where,
			fmt.Sprintf("is_active = $%d", len(args)),
		)
	}

	whereSQL := ""

	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}

	// COUNT(*) dengan kondisi WHERE yang sama
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM students
		%s
	`, whereSQL)

	var total int

	if err := r.pool.QueryRow(
		ctx,
		countQuery,
		args...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Parameter LIMIT dan OFFSET
	queryArgs := append([]any{}, args...)

	queryArgs = append(queryArgs, params.Limit)
	limitParam := len(queryArgs)

	queryArgs = append(queryArgs, params.Offset)
	offsetParam := len(queryArgs)

	// Query data
	query := fmt.Sprintf(`
		SELECT
			id,
			nim,
			name,
			grade,
			is_active,
			created_at
		FROM students
		%s
		ORDER BY %s %s
		LIMIT $%d
		OFFSET $%d
	`,
		whereSQL,
		sortColumn,
		order,
		limitParam,
		offsetParam,
	)

	rows, err := r.pool.Query(
		ctx,
		query,
		queryArgs...,
	)

	if err != nil {
		return nil, 0, err
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
			return nil, 0, err
		}

		students = append(students, s)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return students, total, nil
}

func (r *studentPostgresRepository) FindByID(
	ctx context.Context,
	id int,
) (model.Student, error) {

	var s model.Student

	err := r.pool.QueryRow(
		ctx,
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

	err := r.pool.QueryRow(
		ctx,
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

	err := r.pool.QueryRow(
		ctx,
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