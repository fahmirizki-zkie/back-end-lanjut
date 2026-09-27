package service

import (
	"testing"

	"modul6/app/model"
	"modul6/helper"
)

// Pengujian tidak menyalakan server,
// tidak menyentuh database,
// dan tidak membuat fiber.Ctx.

func TestValidateCreateDeclarative(t *testing.T) {
	req := model.CreateStudentRequest{
		Name:     "",
		Email:    "email-salah",
		NIM:      "123", // kurang dari 9 digit (tag nim)
		Password: "123",
	}

	errs := helper.ValidateStruct(req)
	if len(errs) == 0 {
		t.Errorf("diharapkan error validasi, tapi tidak ada")
	}
	if _, ok := errs["name"]; !ok {
		t.Errorf("diharapkan error pada name")
	}
	if _, ok := errs["email"]; !ok {
		t.Errorf("diharapkan error pada email")
	}
	if _, ok := errs["nim"]; !ok {
		t.Errorf("diharapkan error pada nim")
	}
	if _, ok := errs["password"]; !ok {
		t.Errorf("diharapkan error pada password")
	}
}

func TestValidateReplaceDeclarative(t *testing.T) {
	req := model.ReplaceStudentRequest{
		Name:  "",
		Email: "email-salah",
		NIM:   "abc", // bukan angka
	}

	errs := helper.ValidateStruct(req)
	if len(errs) == 0 {
		t.Errorf("diharapkan error validasi, tapi tidak ada")
	}
	if _, ok := errs["name"]; !ok {
		t.Errorf("diharapkan error pada name")
	}
	if _, ok := errs["email"]; !ok {
		t.Errorf("diharapkan error pada email")
	}
	if _, ok := errs["nim"]; !ok {
		t.Errorf("diharapkan error pada nim")
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

	result := ApplyPatch(
		initial,
		model.UpdateStudentRequest{
			IsActive: &inactive,
		},
	)

	if result.IsActive {
		t.Error("is_active seharusnya berubah menjadi false")
	}

	if result.Name != "Fahmi Rizky" {
		t.Error("field yang tidak dikirim seharusnya tidak berubah")
	}
}
