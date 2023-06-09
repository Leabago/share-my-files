package models

type Product struct {
	// gorm.Model
	Code  string
	Price uint
}

type File struct {
	// gorm.Model
	Name string
}
