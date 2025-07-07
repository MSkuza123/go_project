package data

import (
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

type PDFDocument struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Pdf     []byte `json:"pdf"`
}

type PDFModel struct {
	DB *sql.DB
}

func (p PDFModel) Insert(pdf PDFDocument) error {
	query := `
	INSERT INTO pdf_documents (id, title, content, pdf)
	VALUES ($1, $2, $3, $4)`

	args := []interface{}{pdf.ID, pdf.Title, pdf.Content, pdf.Pdf}

	_, err := p.DB.Exec(query, args...)
	if err != nil {
		return err
	}
	return nil
}

func (p PDFModel) Get(id int64) (*PDFDocument, error) {
	if id < 1 {
		return nil, errors.New("record not found")
	}

	query := `
		SELECT id, title, content, pdf
		FROM pdf_documents
		WHERE id = $1`

	var pdf PDFDocument

	err := p.DB.QueryRow(query, id).Scan(
		&pdf.ID,
		&pdf.Title,
		&pdf.Content,
		&pdf.Pdf,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("record not found")
		default:
			return nil, err
		}
	}

	return &pdf, nil

}

func (p PDFModel) Update(pdf *PDFDocument) error {
	query := `
		UPDATE pdf_documents
		SET title = $1, content = $2, pdf = $3
		WHERE id = $4`

	args := []interface{}{pdf.Title, pdf.Content, pdf.Pdf, pdf.ID}

	_, err := p.DB.Exec(query, args...)
	if err != nil {
		return err
	}
	return nil
}

func (p PDFModel) Delete(id int64) error {
	if id < 1 {
		return errors.New("record not found")
	}

	query := `
		DELETE FROM pdf_documents
		WHERE id = $1`

	results, err := p.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := results.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("record not found")
	}

	return nil
}

func (p PDFModel) GetAll() ([]*PDFDocument, error) {
	query := `
	  SELECT * 
	  FROM pdf_documents
	  ORDER BY id`

	rows, err := p.DB.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	pdfs := []*PDFDocument{}

	for rows.Next() {
		var pdf PDFDocument

		err := rows.Scan(
			&pdf.ID,
			&pdf.Title,
			&pdf.Content,
			&pdf.Pdf,
		)
		if err != nil {
			return nil, err
		}

		pdfs = append(pdfs, &pdf)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return pdfs, nil
}
