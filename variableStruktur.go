package main 

import "fmt"

func main() {

	var name string = "Fahmi"
	var age int = 20
	var tinggi float64 = 179.9
	var isStudent bool = true
	var hobi = []string{"Sepakbola" , "Main PubgMobile", "Badminton"}


	mahasiswa := map[string] int{
		"Azzam"	: 95,
		"Fahmi"	: 98,
		"Ody"	: 90,
		"Raihan": 85,
	} 

	// Menambah 
	mahasiswa["Anshari"] = 93
	
	// Cek dari keberadaan 
	if nilai , oke := mahasiswa["Fahmi"]; oke {
		fmt.Println("Nilai Fahmi adalah", nilai)
	} else {
		fmt.Println("Fahmi tidak ditemukan")
	}

	// Menghapus
	delete(mahasiswa, "Ody")

	// Menelusuri seluruh isi 
	fmt.Println("Daftar Mahasiswa dan Nilainya:")
	for nama, nilai := range mahasiswa {
		fmt.Println(nama, "=", nilai)
	}
}