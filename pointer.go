package main 

import "fmt"


// Pass by value
func swapa(a,b int) {a, b = b, a}

// Pass by pointer 
func swapb(a,b *int) {*a, *b = *b, *a}

// pass by value
func updateA(s []string, newItem string) {s = append(s, newItem)}

// pass by pointer
func updateB(s *[]string, newItem string) {*s = append(*s, newItem)}

func main() {
	
	// menukar nilai pakai pass by value
	a, b := 1, 2
	swapa(a, b)
	fmt.Println("pakai pass by value:", a, b)

	// menukar nilai pakai pointer
	a, b = 1, 2
	swapb(&a, &b)
	fmt.Println("pakai pass by pointer:", a, b)

	// update pakai pass by value
	Hobi := []string{"Bola", "Main Pubg", "Badminton"}
	updateA(Hobi, "Gaming")
	fmt.Println("pakai pass by value:", Hobi) 

	// update pakai pointer
	Hobi = []string{"Bola", "Main Pubg", "Badminton"}
	updateB(&Hobi, "Gaming")
	fmt.Println("pakai pass by pointer:", Hobi) 
}