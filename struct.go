package main 

import "fmt"

type User struct {
 ID int 
 Name string 
 Grade float64
 IsActive bool 
}

// Method untuk mendapatkan informasi user
func (u User) GetInfo() string {
	return fmt.Sprintf("ID: %d, Name: %s, Grade: %.2f, Active: %t", u.ID, u.Name, u.Grade, u.IsActive)
}

// Method untuk mengupdate grade user
func (u *User) UpdateGrade(Grade float64) {
	u.Grade = Grade
} 

// Method untuk mengupdate Status user
func (u *User) Activate() { 
	u.IsActive = true 
}
func (u *User) Deactivate() { 
	u.IsActive = false 
}
func main() {
	// Daftar Mahasiswa
	mahasiswa := []User{{ID: 1, Name: "Fahmi", Grade: 3.6, IsActive: false}}

	// Menampilkan informasi mahasiswa
	fmt.Println("Menampilkan informasi mahasiswa:", mahasiswa[0].GetInfo())

	// Memperbarui nilai
	mahasiswa[0].UpdateGrade(4.0)
	fmt.Println("Memperbarui nilai:", mahasiswa[0].GetInfo())

	// Mengubah status aktif
	mahasiswa[0].Activate()
	fmt.Println("Mengubah status aktif:", mahasiswa[0].GetInfo())

	// Mengubah status non-aktif
	mahasiswa[0].Deactivate()
	fmt.Println("Mengubah status non-aktif:", mahasiswa[0].GetInfo())
}	