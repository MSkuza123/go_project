package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"pdfcrudsign/internal/data"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func setupTestApp(t *testing.T) *application {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	createTableSQL := `
	CREATE TABLE pdf_documents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		pdf BLOB NOT NULL
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		t.Fatal(err)
	}

	models := data.NewModels(db)

	return &application{
		models: models,
	}
}

func insertOrUpdatePDFHandlerMock(app *application) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")

		var input struct {
			ID      int64  `json:"id"`
			Title   string `json:"title"`
			Content string `json:"content"`
		}

		if err := c.BindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid inputs"})
			return
		}

		if idParam == "" {
			doc := data.PDFDocument{
				ID:      input.ID,
				Title:   input.Title,
				Content: input.Content,
				Pdf:     []byte("dummy pdf bytes"),
			}

			if err := app.models.PDFDocuments.Insert(doc); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database insert failed"})
				return
			}

			c.JSON(http.StatusCreated, gin.H{"message": "PDF created and stored"})
			return
		}

		id, err := strconv.ParseInt(idParam, 10, 64)
		doc, err := app.models.PDFDocuments.Get(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "PDF not found"})
			return
		}

		doc.Title = input.Title
		doc.Content = input.Content
		doc.Pdf = []byte("dummy pdf bytes")

		if err := app.models.PDFDocuments.Update(doc); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database update failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":      doc.ID,
			"title":   doc.Title,
			"content": doc.Content,
			"pdf":     doc.Pdf,
		})
	}
}

func setupRouter(app *application) *gin.Engine {
	router := gin.Default()
	router.POST("/pdfs", insertOrUpdatePDFHandlerMock(app))
	router.PUT("/pdfs/:id", insertOrUpdatePDFHandlerMock(app))
	router.GET("/pdfs/:id", app.getPDFHandler)
	router.GET("/pdfs", app.getAllPDFsHandler)
	router.DELETE("/pdfs/:id", app.deletePDFHandler)
	return router
}

func TestInsertAndGetPDF(t *testing.T) {
	app := setupTestApp(t)
	router := setupRouter(app)

	payload := map[string]interface{}{
		"id":      1,
		"title":   "Test PDF",
		"content": "Test content",
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/pdfs", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/pdfs/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		ID      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
		Pdf     []byte `json:"pdf"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), response.ID)
	assert.Equal(t, "Test PDF", response.Title)
	assert.Equal(t, "Test content", response.Content)
	assert.Equal(t, []byte("dummy pdf bytes"), response.Pdf)
}

func TestUpdatePDF(t *testing.T) {
	app := setupTestApp(t)
	router := setupRouter(app)

	app.models.PDFDocuments.Insert(data.PDFDocument{
		ID:      1,
		Title:   "Old Title",
		Content: "Old Content",
		Pdf:     []byte("dummy pdf bytes"),
	})

	payload := map[string]interface{}{
		"title":   "New Title",
		"content": "New Content",
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/pdfs/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		ID      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
		Pdf     []byte `json:"pdf"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), response.ID)
	assert.Equal(t, "New Title", response.Title)
	assert.Equal(t, "New Content", response.Content)
	assert.Equal(t, []byte("dummy pdf bytes"), response.Pdf)
}

func TestGetAllPDFs(t *testing.T) {
	app := setupTestApp(t)
	router := setupRouter(app)

	app.models.PDFDocuments.Insert(data.PDFDocument{
		ID:      1,
		Title:   "PDF 1",
		Content: "Content 1",
		Pdf:     []byte("dummy"),
	})
	app.models.PDFDocuments.Insert(data.PDFDocument{
		ID:      2,
		Title:   "PDF 2",
		Content: "Content 2",
		Pdf:     []byte("dummy"),
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pdfs", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "PDF 1")
	assert.Contains(t, w.Body.String(), "PDF 2")
}

func TestDeletePDF(t *testing.T) {
	app := setupTestApp(t)
	router := setupRouter(app)

	if _, err := os.Stat("C:/Users/mskuza/OneDrive - Capgemini/Desktop/Go PDF/cmd/api/ARIAL.TTF"); os.IsNotExist(err) {
		t.Fatal("Lack of font ./arial/ARIAL.TTF")
	}
	if _, err := os.Stat("C:/Users/mskuza/OneDrive - Capgemini/Desktop/Go PDF/cert.p12"); os.IsNotExist(err) {
		t.Fatal("Lack of certificate cert.p12")
	}

	app.models.PDFDocuments.Insert(data.PDFDocument{
		ID:      1,
		Title:   "ToDelete",
		Content: "ToDelete",
		Pdf:     []byte("dummy"),
	})

	_ = os.WriteFile("1.pdf", []byte("dummy"), 0644)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/pdfs/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "PDF deleted")
}
