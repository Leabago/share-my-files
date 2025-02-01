package operation

import (
	"share-my-files/pkg/models"

	"gorm.io/gorm"
)

type FileModel struct {
	DB *gorm.DB
}

func (file *FileModel) Insert() {
	file.DB.Create(&models.File{})
}

func (file *FileModel) Get() {

}
func (file *FileModel) Latest() {

}
