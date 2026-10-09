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

type CourseRepository interface {
	FindAll(ctx context.Context, semester string, search string, availableOnly bool) ([]model.CourseResponse, error)
	FindByID(ctx context.Context, id int64) (*model.Course, error)
	GetForUpdate(ctx context.Context, tx pgx.Tx, id int64) (*model.Course, error)
	CountEnrolled(ctx context.Context, tx pgx.Tx, courseID int64, tahunAkademik string) (int, error)
}

type courseRepository struct {
	pool *pgxpool.Pool
}

func NewCourseRepository(pool *pgxpool.Pool) CourseRepository {
	return &courseRepository{pool: pool}
}

func (r *courseRepository) FindAll(ctx context.Context, semester string, search string, availableOnly bool) ([]model.CourseResponse, error) {
	whereClauses := []string{"1=1"}
	args := []any{}
	argIdx := 1

	if strings.TrimSpace(semester) != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("c.semester = $%d", argIdx))
		args = append(args, semester)
		argIdx++
	}

	if strings.TrimSpace(search) != "" {
		pattern := "%" + strings.ToLower(search) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("(LOWER(c.kode_mk) LIKE $%d OR LOWER(c.nama_mk) LIKE $%d)", argIdx, argIdx))
		args = append(args, pattern)
		argIdx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	query := fmt.Sprintf(`
		SELECT 
			c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota,
			COALESCE(COUNT(e.id), 0) AS terisi,
			(c.kuota - COALESCE(COUNT(e.id), 0)) AS sisa_kuota
		FROM courses c
		LEFT JOIN enrollments e ON e.course_id = c.id
		WHERE %s
		GROUP BY c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota
	`, whereSQL)

	if availableOnly {
		query += " HAVING (c.kuota - COALESCE(COUNT(e.id), 0)) > 0"
	}

	query += " ORDER BY c.semester ASC, c.kode_mk ASC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	courses := make([]model.CourseResponse, 0)
	for rows.Next() {
		var item model.CourseResponse
		if err := rows.Scan(&item.ID, &item.KodeMK, &item.NamaMK, &item.SKS, &item.Semester, &item.Kuota, &item.Terisi, &item.SisaKuota); err != nil {
			return nil, err
		}
		courses = append(courses, item)
	}

	return courses, nil
}

func (r *courseRepository) FindByID(ctx context.Context, id int64) (*model.Course, error) {
	query := `
		SELECT id, kode_mk, nama_mk, sks, semester, kuota, created_at, updated_at
		FROM courses
		WHERE id = $1
		LIMIT 1
	`
	row := r.pool.QueryRow(ctx, query, id)

	var c model.Course
	err := row.Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// GetForUpdate mengunci baris mata kuliah dengan SELECT ... FOR UPDATE
func (r *courseRepository) GetForUpdate(ctx context.Context, tx pgx.Tx, id int64) (*model.Course, error) {
	query := `
		SELECT id, kode_mk, nama_mk, sks, semester, kuota, created_at, updated_at
		FROM courses
		WHERE id = $1
		FOR UPDATE
	`
	row := tx.QueryRow(ctx, query, id)

	var c model.Course
	err := row.Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *courseRepository) CountEnrolled(ctx context.Context, tx pgx.Tx, courseID int64, tahunAkademik string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM enrollments
		WHERE course_id = $1 AND tahun_akademik = $2
	`
	var count int
	var err error
	if tx != nil {
		err = tx.QueryRow(ctx, query, courseID, tahunAkademik).Scan(&count)
	} else {
		err = r.pool.QueryRow(ctx, query, courseID, tahunAkademik).Scan(&count)
	}
	return count, err
}
