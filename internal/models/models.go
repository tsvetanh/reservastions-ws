package models

type ModelList []interface{}

var AllModels = ModelList{
	User{},
	UserRoles{},
	Role{},
	Hall{},
	HallImage{},
	Reservation{},
	RefreshToken{},
}
