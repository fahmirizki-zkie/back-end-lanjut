package service

import (
	"strings"

	"modul6/app/model"
)

// File ini berisi business rules MURNI: tidak menyentuh fiber.Ctx,
// tidak menyentuh database, dan tidak tahu apa pun tentang HTTP.

// ValidateCreate memeriksa isi permintaan pembuatan student.
// Mengembalikan peta berisi field yang bermasalah; kosong berarti lolos.
func ValidateCreate(req model.CreateStudentRequest) map[string]string {
	errs := map[string]string{}

	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi"
	}

	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi"
	}

	if !isValidEmail(req.Email) {
		errs["email"] = "format email tidak valid"
	}

	if len(req.Password) < 8 {
		errs["password"] = "minimal 8 karakter"
	}

	return errs
}

// ValidateReplace memeriksa isi permintaan PUT.
// Seluruh field wajib ada karena PUT mengganti isi secara keseluruhan.
func ValidateReplace(req model.ReplaceStudentRequest) map[string]string {
	errs := map[string]string{}

	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi pada PUT"
	}

	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi pada PUT"
	}

	if !isValidEmail(req.Email) {
		errs["email"] = "wajib diisi dan berformat email pada PUT"
	}

	return errs
}

// ApplyPatch menyalin field yang dikirim ke data yang sudah ada.
// Field yang bernilai nil dibiarkan apa adanya.
func ApplyPatch(
	current model.Student, req model.UpdateStudentRequest,
) (model.Student, map[string]string) {
	errs := map[string]string{}

	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			errs["name"] = "tidak boleh kosong"
		} else {
			current.Name = strings.TrimSpace(*req.Name)
		}
	}

	if req.Email != nil {
		if !isValidEmail(*req.Email) {
			errs["email"] = "format email tidak valid"
		} else {
			current.Email = strings.TrimSpace(*req.Email)
		}
	}

	if req.NIM != nil {
		if strings.TrimSpace(*req.NIM) == "" {
			errs["nim"] = "tidak boleh kosong"
		} else {
			current.NIM = strings.TrimSpace(*req.NIM)
		}
	}

	if req.Grade != nil {
		current.Grade = *req.Grade
	}

	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}

	return current, errs
}

// IsEmptyPatch menandai permintaan PATCH yang tidak mengubah apa pun.
func IsEmptyPatch(req model.UpdateStudentRequest) bool {
	return req.Name == nil &&
		req.Email == nil &&
		req.NIM == nil &&
		req.Grade == nil &&
		req.IsActive == nil
}

// CountTotalPages membulatkan ke atas tanpa memakai bilangan pecahan.
func CountTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}

	return (total + limit - 1) / limit
}

// isValidEmail adalah pemeriksaan sederhana, bukan validasi RFC.
func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	at := strings.Index(email, "@")
	dot := strings.LastIndex(email, ".")

	return at > 0 && dot > at+1 && dot < len(email)-1
}
