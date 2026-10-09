package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/model"
)

type EnrollmentRepository interface {
	CheckDuplicate(ctx context.Context, tx pgx.Tx, studentID, courseID int64, tahunAkademik string) (bool, error)
	GetTotalSKS(ctx context.Context, tx pgx.Tx, studentID int64, tahunAkademik string) (int, error)
	Create(ctx context.Context, tx pgx.Tx, enrollment *model.Enrollment) error
	FindByID(ctx context.Context, id int64) (*model.Enrollment, error)
	Delete(ctx context.Context, id int64) error
	GetCoursesByStudentID(ctx context.Context, studentID int64) ([]model.EnrolledCourseItem, error)
	GetTotalSKSByStudentID(ctx context.Context, studentID int64) (int, error)
}

type enrollmentRepository struct {
	pool *pgxpool.Pool
}

func NewEnrollmentRepository(pool *pgxpool.Pool) EnrollmentRepository {
	return &enrollmentRepository{pool: pool}
}

func (r *enrollmentRepository) CheckDuplicate(ctx context.Context, tx pgx.Tx, studentID, courseID int64, tahunAkademik string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM enrollments 
			WHERE student_id = $1 AND course_id = $2 AND tahun_akademik = $3
		)
	`
	var exists bool
	var err error
	if tx != nil {
		err = tx.QueryRow(ctx, query, studentID, courseID, tahunAkademik).Scan(&exists)
	} else {
		err = r.pool.QueryRow(ctx, query, studentID, courseID, tahunAkademik).Scan(&exists)
	}
	return exists, err
}

func (r *enrollmentRepository) GetTotalSKS(ctx context.Context, tx pgx.Tx, studentID int64, tahunAkademik string) (int, error) {
	query := `
		SELECT COALESCE(SUM(c.sks), 0)
		FROM enrollments e
		JOIN courses c ON c.id = e.course_id
		WHERE e.student_id = $1 AND e.tahun_akademik = $2
	`
	var totalSKS int
	var err error
	if tx != nil {
		err = tx.QueryRow(ctx, query, studentID, tahunAkademik).Scan(&totalSKS)
	} else {
		err = r.pool.QueryRow(ctx, query, studentID, tahunAkademik).Scan(&totalSKS)
	}
	return totalSKS, err
}

func (r *enrollmentRepository) GetTotalSKSByStudentID(ctx context.Context, studentID int64) (int, error) {
	query := `
		SELECT COALESCE(SUM(c.sks), 0)
		FROM enrollments e
		JOIN courses c ON c.id = e.course_id
		WHERE e.student_id = $1
	`
	var totalSKS int
	err := r.pool.QueryRow(ctx, query, studentID).Scan(&totalSKS)
	return totalSKS, err
}

func (r *enrollmentRepository) Create(ctx context.Context, tx pgx.Tx, enrollment *model.Enrollment) error {
	query := `
		INSERT INTO enrollments (student_id, course_id, tahun_akademik, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, created_at
	`
	return tx.QueryRow(ctx, query, enrollment.StudentID, enrollment.CourseID, enrollment.TahunAkademik).
		Scan(&enrollment.ID, &enrollment.CreatedAt)
}

func (r *enrollmentRepository) FindByID(ctx context.Context, id int64) (*model.Enrollment, error) {
	query := `
		SELECT id, student_id, course_id, tahun_akademik, created_at
		FROM enrollments
		WHERE id = $1
		LIMIT 1
	`
	row := r.pool.QueryRow(ctx, query, id)

	var e model.Enrollment
	err := row.Scan(&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r *enrollmentRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM enrollments WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *enrollmentRepository) GetCoursesByStudentID(ctx context.Context, studentID int64) ([]model.EnrolledCourseItem, error) {
	query := `
		SELECT e.id, c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, e.tahun_akademik
		FROM enrollments e
		JOIN courses c ON c.id = e.course_id
		WHERE e.student_id = $1
		ORDER BY e.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	courses := make([]model.EnrolledCourseItem, 0)
	for rows.Next() {
		var item model.EnrolledCourseItem
		if err := rows.Scan(&item.EnrollmentID, &item.CourseID, &item.KodeMK, &item.NamaMK, &item.SKS, &item.Semester, &item.TahunAkademik); err != nil {
			return nil, err
		}
		courses = append(courses, item)
	}
	return courses, nil
}
