package service

import (
	"strings"

	"modul6/app/model"
)

// File ini berisi business rules MURNI: tidak menyentuh fiber.Ctx,
// tidak menyentuh database, dan tidak tahu apa pun tentang HTTP.
// Seluruh validasi format field sudah dipindahkan ke tag deklaratif pada struct.

// ApplyPatch menyalin field yang dikirim ke data yang sudah ada.
// Pemeriksaan format sudah ditangani oleh tag validator.
func ApplyPatch(
	current model.Student, req model.UpdateStudentRequest,
) model.Student {
	if req.Name != nil {
		current.Name = strings.TrimSpace(*req.Name)
	}
	if req.Email != nil {
		current.Email = strings.TrimSpace(*req.Email)
	}
	if req.NIM != nil {
		current.NIM = strings.TrimSpace(*req.NIM)
	}
	if req.Grade != nil {
		current.Grade = *req.Grade
	}
	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}

	return current
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
