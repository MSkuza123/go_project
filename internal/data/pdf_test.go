package data

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestPDFModel_Insert(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	model := PDFModel{DB: db}

	doc := PDFDocument{ID: 1, Title: "Title", Content: "Content", Pdf: []byte("PDF")}

	mock.ExpectExec(`INSERT INTO pdf_documents`).
		WithArgs(doc.ID, doc.Title, doc.Content, doc.Pdf).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := model.Insert(doc)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

}

func TestPDFModel_Get(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	model := PDFModel{DB: db}

	expected := PDFDocument{ID: 2, Title: "Doc", Content: "Cont", Pdf: []byte("abc")}
	rows := sqlmock.NewRows([]string{"id", "title", "content", "pdf"}).
		AddRow(expected.ID, expected.Title, expected.Content, expected.Pdf)

	mock.ExpectQuery(`SELECT id, title, content, pdf FROM pdf_documents WHERE id = \$1`).
		WithArgs(expected.ID).
		WillReturnRows(rows)

	got, err := model.Get(expected.ID)
	assert.NoError(t, err)
	assert.Equal(t, expected, *got)
}

func TestPDFModel_Update(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	model := PDFModel{DB: db}

	doc := &PDFDocument{ID: 1, Title: "New", Content: "Content", Pdf: []byte("updated")}
	mock.ExpectExec(`UPDATE pdf_documents SET title = \$1, content = \$2, pdf = \$3 WHERE id = \$4`).
		WithArgs(doc.Title, doc.Content, doc.Pdf, doc.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := model.Update(doc)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPDFModel_Delete(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	model := PDFModel{DB: db}
	id := int64(1)

	mock.ExpectExec(`DELETE FROM pdf_documents WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := model.Delete(id)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
