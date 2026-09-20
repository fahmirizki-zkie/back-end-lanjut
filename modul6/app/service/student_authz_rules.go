package service
 
import (
    "strings"
 
    "latihan-fiber/app/model"
    "latihan-fiber/helper"
)
 
// CanAccessUser memutuskan apakah seseorang boleh menyentuh data user lain.
//
// Dua jalur yang diizinkan:
//  1. Kepemilikan (ownership) — data itu miliknya sendiri.
//  2. Permission — role-nya memang berhak atas data siapa pun.
//
// Urutannya disengaja: pemeriksaan kepemilikan didahulukan karena paling
// murah dan paling sering benar. Bila keduanya gagal, jawabannya false.
func CanAccessStudent(
	current model.AuthUser,
	ownerID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	if current.UserID == ownerID {
		return true
	}

	return perms.Can(current.Role, anyPermission)
}
 
