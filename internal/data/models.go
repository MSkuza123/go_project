package data

import "database/sql"

type Models struct {
	PDFDocuments PDFModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		PDFDocuments: PDFModel{DB: db},
	}
}
