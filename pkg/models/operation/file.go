package operation

import (
	"share-my-file/pkg/models"

	"gorm.io/gorm"
)

type FileModel struct {
	DB *gorm.DB
}

func (file *FileModel) Insert() {
	file.DB.Create(&models.File{Name: "file1"})
}

func (file *FileModel) Get() {

}
func (file *FileModel) Latest() {

}
