package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/model"
)

type StudentRepository interface {
	FindAll(ctx context.Context, page, perPage int, prodi, angkatan string, search, sort string) ([]model.StudentItem, int64, error)
	FindByID(ctx context.Context, id int64) (*model.Student, error)
	FindByUserID(ctx context.Context, userID int64) (*model.Student, error)
	CheckNIMExists(ctx context.Context, nim string) (bool, error)
	Create(ctx context.Context, tx pgx.Tx, student *model.Student) error
	Update(ctx context.Context, student *model.Student) error
	SoftDelete(ctx context.Context, id int64) error
}

type studentRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentRepository{pool: pool}
}

func (r *studentRepository) FindAll(ctx context.Context, page, perPage int, prodi, angkatan string, search, sort string) ([]model.StudentItem, int64, error) {
	whereClauses := []string{"deleted_at IS NULL"}
	args := []any{}
	argIdx := 1

	if strings.TrimSpace(prodi) != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("LOWER(prodi) = LOWER($%d)", argIdx))
		args = append(args, prodi)
		argIdx++
	}

	if strings.TrimSpace(angkatan) != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("angkatan = $%d", argIdx))
		args = append(args, angkatan)
		argIdx++
	}

	if strings.TrimSpace(search) != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("(LOWER(nama) LIKE $%d OR LOWER(nim) LIKE $%d)", argIdx, argIdx))
		args = append(args, searchPattern)
		argIdx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// Hitung total data
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM students WHERE %s", whereSQL)
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Atur sorting
	orderBy := "id ASC"
	switch sort {
	case "nama":
		orderBy = "nama ASC"
	case "-nama":
		orderBy = "nama DESC"
	case "ipk_terakhir":
		orderBy = "ipk_terakhir ASC"
	case "-ipk_terakhir":
		orderBy = "ipk_terakhir DESC"
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT id, nim, nama, prodi, angkatan, ipk_terakhir
		FROM students
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereSQL, orderBy, argIdx, argIdx+1)

	queryArgs := append(args, perPage, offset)

	rows, err := r.pool.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.StudentItem, 0)
	for rows.Next() {
		var s model.StudentItem
		if err := rows.Scan(&s.ID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir); err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}

	return items, total, nil
}

func (r *studentRepository) FindByID(ctx context.Context, id int64) (*model.Student, error) {
	query := `
		SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at, created_at, updated_at
		FROM students
		WHERE id = $1
		LIMIT 1
	`
	row := r.pool.QueryRow(ctx, query, id)

	var s model.Student
	err := row.Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *studentRepository) FindByUserID(ctx context.Context, userID int64) (*model.Student, error) {
	query := `
		SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at, created_at, updated_at
		FROM students
		WHERE user_id = $1
		LIMIT 1
	`
	row := r.pool.QueryRow(ctx, query, userID)

	var s model.Student
	err := row.Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *studentRepository) CheckNIMExists(ctx context.Context, nim string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM students WHERE nim = $1)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, nim).Scan(&exists)
	return exists, err
}

func (r *studentRepository) Create(ctx context.Context, tx pgx.Tx, student *model.Student) error {
	query := `
		INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	var err error
	if tx != nil {
		err = tx.QueryRow(ctx, query, student.UserID, student.NIM, student.Nama, student.Prodi, student.Angkatan, student.IPKTerakhir).
			Scan(&student.ID, &student.CreatedAt, &student.UpdatedAt)
	} else {
		err = r.pool.QueryRow(ctx, query, student.UserID, student.NIM, student.Nama, student.Prodi, student.Angkatan, student.IPKTerakhir).
			Scan(&student.ID, &student.CreatedAt, &student.UpdatedAt)
	}
	return err
}

func (r *studentRepository) Update(ctx context.Context, student *model.Student) error {
	query := `
		UPDATE students
		SET nama = $1, prodi = $2, angkatan = $3, ipk_terakhir = $4, updated_at = NOW()
		WHERE id = $5 AND deleted_at IS NULL
		RETURNING updated_at
	`
	return r.pool.QueryRow(ctx, query, student.Nama, student.Prodi, student.Angkatan, student.IPKTerakhir, student.ID).
		Scan(&student.UpdatedAt)
}

func (r *studentRepository) SoftDelete(ctx context.Context, id int64) error {
	query := `
		UPDATE students
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
