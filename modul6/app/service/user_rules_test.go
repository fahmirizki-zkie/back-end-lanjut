package service

import (
	"testing"

	"modul6/app/model"
)

// Pengujian tidak menyalakan server,
// tidak menyentuh database,
// dan tidak membuat fiber.Ctx.

func TestValidateCreate(t *testing.T) {
	req := model.CreateStudentRequest{
		Name:     "",
		Email:    "email-salah",
		NIM:      "",
		Password: "123",
	}

	errs := ValidateCreate(req)

	if len(errs) != 4 {
		t.Errorf("harap 4 error validasi, dapat %d: %v", len(errs), errs)
	}
}

func TestValidateReplace(t *testing.T) {
	req := model.ReplaceStudentRequest{
		Name:  "",
		Email: "email-salah",
		NIM:   "",
	}

	errs := ValidateReplace(req)

	if len(errs) != 3 {
		t.Errorf("harap 3 error validasi, dapat %d: %v", len(errs), errs)
	}
}

func TestApplyPatch(t *testing.T) {
	initial := model.Student{
		ID:       1,
		Name:     "Fahmi Rizky",
		Email:    "fahmi@gmail.com",
		NIM:      "434241016",
		Grade:    85,
		IsActive: true,
	}

	inactive := false

	result, errs := ApplyPatch(
		initial,
		model.UpdateStudentRequest{
			IsActive: &inactive,
		},
	)

	if len(errs) != 0 {
		t.Fatalf("tidak seharusnya ada error: %v", errs)
	}

	if result.IsActive {
		t.Error("is_active seharusnya berubah menjadi false")
	}

	if result.Name != "Fahmi Rizky" {
		t.Error("field yang tidak dikirim seharusnya tidak berubah")
	}
}