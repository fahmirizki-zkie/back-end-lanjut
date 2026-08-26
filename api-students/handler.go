package main 

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

var students []Student

var nextStudentID = 1

// fungsi mencari student 
func findstudentindex(id int) int {
	for i := range students {
		if students[i].ID == id {
			return i
		}
	}
	return -1
}

func paramStudentID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))

	if err != nil || id < 1 {
		return 0, false
	}

	return id, true
}

func getStudent(c *fiber.Ctx) error {
	id, valid := paramStudentID(c)

	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")

	}
	
	i := findstudentindex(id)

	if i == -1 {
		return fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	}

	return ok(c, "student ditemukan", students[i])
}

func createStudent(c *fiber.Ctx) error {
	var req CreateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, " body harus berupa json yang valid")
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.NIM = strings.TrimSpace(req.NIM)

	errs := map[string]string{}

	if req.Name == "" {
		errs["name"] = "wajib diisi"
	}

	if req.Email == "" {
		errs["email"] = "wajib diisi"
	}
	
	if req.NIM == "" {
		errs["nim"] = "wajib diisi"
	}

	if req.Password == "" {
		errs["password"] = "wajib diisi"
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	//cek nim unik 
	for _, student := range students {
		if strings.EqualFold(student.NIM, req.NIM) {
			return fail(
				c,fiber.StatusConflict, "nim sudah digunakan",
			)
		}
	}

	newStudent := Student{
		ID:        nextStudentID,
		Name:      req.Name,
		Email:     req.Email,
		NIM:       req.NIM,
		Password:  req.Password,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	students = append(students, newStudent)
	nextStudentID++

	return created(
		c,"student berhasil dibuat", newStudent, "/students/"+strconv.Itoa(newStudent.ID),
	)
}

func replaceStudent(c *fiber.Ctx) error {
	id, valid := paramStudentID(c)

	if !valid {
		return fail(
			c, fiber.StatusBadRequest, "id harus berupa angka positif",
		)
	}

	i := findstudentindex(id)

	if i == -1 {
		return fail(
			c, fiber.StatusNotFound, "student tidak ditemukan",
		)
	}

	var req ReplaceStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(
			c, fiber.StatusBadRequest, "body harus berupa json yang valid",
		)
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.NIM = strings.TrimSpace(req.NIM)

	errs := map[string]string{}

	if req.Name == "" {
		errs["name"] = "wajib diisi"
	}

	if req.Email == "" {
		errs["email"] = "wajib diisi"
	}

	if req.NIM == "" {
		errs["nim"] = "wajib diisi"
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	// cek apakah nim dipakai student lain
	for j, student := range students {
		if j != i && strings.EqualFold(student.NIM, req.NIM) {
			return fail(
				c, fiber.StatusConflict, "nim sudah digunakan",
			)
		}
	}

	students[i].Name = req.Name
	students[i].Email = req.Email
	students[i].NIM = req.NIM
	students[i].IsActive = req.IsActive

	return ok(
		c, "student berhasil diubah", students[i],
	)
}

func patchStudent(c *fiber.Ctx) error {
	id, valid := paramStudentID(c)

	if !valid {
		return fail(
			c, fiber.StatusBadRequest, "id harus berupa angka positif",
		)
	}

	i := findstudentindex(id)

	if i == -1 {
		return fail(
			c, fiber.StatusNotFound, "student tidak ditemukan",
		)
	}

	var req UpdateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(
			c, fiber.StatusBadRequest, "body harus berupa json yang valid",
		)
	}

	if req.Name == nil &&
		req.Email == nil &&
		req.NIM == nil &&
		req.IsActive == nil {
		return fail(
			c, fiber.StatusBadRequest, "tidak ada field yang dikirim",
		)
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)

		if name == "" {
			return failValidation(c, map[string]string{
				"name": "tidak boleh kosong",
			})
		}

		students[i].Name = name	
	}

	if req.Email != nil {
		email := strings.TrimSpace(*req.Email)

		if email == "" {
			return failValidation(c, map[string]string{
				"email": "tidak boleh kosong",
			})
		}

		students[i].Email = email
	}

	if req.NIM != nil {
		nim := strings.TrimSpace(*req.NIM)

		if nim == "" {
			return failValidation(c, map[string]string{
				"nim": "tidak boleh kosong",
			})
		}

		for j, student := range students {
			if j != i && strings.EqualFold(student.NIM, nim){
				return fail(
					c, fiber.StatusConflict, "nim sudah digunakan",
				)
			}
		}

		students[i].NIM = nim
	}

	if req.IsActive != nil {
		students[i].IsActive = *req.IsActive
	}

	return ok(
		c, "student berhasil diubah", students[i],
	)
}

func deleteStudent(c *fiber.Ctx) error {
	id, valid := paramStudentID(c)

	if !valid {
		return fail(
			c, fiber.StatusBadRequest, "id harus berupa angka positif",
		)
	}

	i := findstudentindex(id)

	if i == -1 {
		return fail(
			c, fiber.StatusNotFound, "student tidak ditemukan",
		)
	}

	students = append(
		students[:i], 
		students[i+1:]...)

	return noContent(c)
}

func listStudent(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	search := strings.TrimSpace(c.Query("search"))
	sortField := c.Query("sort", "id")
	order := strings.ToLower(c.Query("order", "asc"))

	isActiveParam := c.Query("is_active")

	var isActive *bool

	if isActiveParam != "" {
		value, err := strconv.ParseBool(isActiveParam)

		if err != nil {
			return fail(
				c, fiber.StatusBadRequest, "is_active harus true atau false",
			)
		}
		isActive = &value
	}

	// salin data agar proses sorting tidak mengubah urutan data utama 
	result := append([]Student(nil), students...)

	//search berdasarkan nama
	if search != "" {
		filtered := []Student{}

		for _, student := range result {
			if strings.Contains(
				strings.ToLower(student.Name), 
				strings.ToLower(search),
			) {
				filtered = append(filtered, student)
			}
		}

		result = filtered
	}

	//filter berdasarkan is_active
	if isActive != nil {
		filtered := []Student{}

		for _, student := range result {
			if student.IsActive == *isActive {
				filtered = append(filtered, student)
			}
		}

		result = filtered
	}

	//sorting whitelist field
	allowedSort := map[string]bool{
		"id": true,
		"name": true,
		"email": true,
		"nim": true,
		"is_active": true,
		"created_at": true,
	}

	if !allowedSort[sortField] {
		return fail(
			c, fiber.StatusBadRequest, "field sort tidak valid",
		)
	}

	if order != "asc" && order != "desc" {
		return fail(
			c, fiber.StatusBadRequest, "order harus asc atau desc",
		)
	}

	sort.Slice(result, func(i, j int) bool {
		var less bool

		switch sortField {
		case "id":
			less = result[i].ID < result[j].ID

		case "name":
			less = strings.ToLower(result[i].Name) < 
			strings.ToLower(result[j].Name)
		
		case "email":
			less = strings.ToLower(result[i].Email) < 
			strings.ToLower(result[j].Email)

		case "nim":
			less = result[i].NIM < result[j].NIM

		case "is_active":
			less = !result[i].IsActive && result[j].IsActive

		case "created_at":
			less = result[i].CreatedAt.Before(
				result[j].CreatedAt,
			)
		}

		if order == "desc" {
			less = !less
		}

		return less
	})

	total := len(result)

	//pagination
	start := (page - 1) * limit

	if start > total {
		start = total
	}

	end := start + limit

	if end > total {
		end = total
	}

	paged := result[start:end]

	totalPage := 0

	if total > 0 {
		totalPage = (total + limit - 1) / limit
	}

	return oklist(
		c, "daftar student berhasil diambil", 
		paged,
		&Meta{
			Page: page,
			Limit: limit,
			Total: total,
			TotalPage: totalPage,
		},
	)
}