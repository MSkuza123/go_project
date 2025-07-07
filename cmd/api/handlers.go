package main

import (
	"crypto"
	"crypto/rsa"
	"fmt"
	"net/http"
	"os"
	"pdfcrudsign/internal/data"
	"strconv"

	"github.com/digitorus/pdfsign/sign"
	"github.com/gin-gonic/gin"
	"github.com/signintech/gopdf"
	"software.sslmate.com/src/go-pkcs12"
)

func (app *application) insertPDFHandler(c *gin.Context) {
	var input struct {
		ID      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid inputs"})
		return
	}

	filename := fmt.Sprintf("%d.pdf", input.ID)
	doc := data.PDFDocument{
		ID:      input.ID,
		Title:   input.Title,
		Content: input.Content,
	}

	if err := generatePDF(doc, filename); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF generation failed"})
		return
	}

	pdfBytes, err := os.ReadFile(filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read generated PDF"})
		return
	}

	doc.Pdf = pdfBytes

	if err := app.models.PDFDocuments.Insert(doc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database insert failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "PDF created and stored"})
}

func (app *application) getPDFHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	doc, err := app.models.PDFDocuments.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PDF not found"})
		return
	}
	type PDFResponse struct {
		ID      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
		Pdf     []byte `json:"pdf"`
	}

	c.IndentedJSON(http.StatusOK, PDFResponse{
		ID:      doc.ID,
		Title:   doc.Title,
		Content: doc.Content,
		Pdf:     doc.Pdf,
	})
}

func (app *application) updatePDFHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	pdf, err := app.models.PDFDocuments.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var input struct {
		Title   *string `json:"title"`
		Content *string `json:"content"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if input.Title != nil {
		pdf.Title = *input.Title
	}
	if input.Content != nil {
		pdf.Content = *input.Content
	}

	filename := fmt.Sprintf("%d.pdf", pdf.ID)
	if err := generatePDF(*pdf, filename); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF generation failed"})
		return
	}

	pdfBytes, err := os.ReadFile(filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read regenerated PDF"})
		return
	}
	pdf.Pdf = pdfBytes

	err = app.models.PDFDocuments.Update(pdf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pdf": pdf})
}

func (app *application) deletePDFHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := app.models.PDFDocuments.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PDF not found or could not be deleted"})
		return
	}

	os.Remove(fmt.Sprintf("%d.pdf", id))
	c.JSON(http.StatusOK, gin.H{"message": "PDF deleted"})
}

func (app *application) getAllPDFsHandler(c *gin.Context) {
	pdfs, err := app.models.PDFDocuments.GetAll()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch PDFs"})
		return
	}

	type PDFInfo struct {
		ID      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	var result []PDFInfo
	for _, pdf := range pdfs {
		result = append(result, PDFInfo{
			ID:      pdf.ID,
			Title:   pdf.Title,
			Content: pdf.Content,
		})
	}

	c.IndentedJSON(http.StatusOK, result)
}

func generatePDF(doc data.PDFDocument, filename string) error {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()

	err := pdf.AddTTFFont("ARIAL", "./arial/ARIAL.TTF")
	if err != nil {
		fmt.Println("Font load error:", err)
		return err
	}

	err = pdf.SetFont("ARIAL", "", 24)
	if err != nil {
		return err
	}
	pdf.SetTextColor(0, 0, 0)
	pdf.SetXY(40, 40)
	pdf.Cell(nil, doc.Title)

	err = pdf.SetFont("ARIAL", "", 14)
	if err != nil {
		return err
	}
	pdf.SetTextColor(0, 0, 0)

	leftMargin := 40.0
	rightMargin := 40.0
	topMargin := 80.0
	bottomMargin := 40.0

	pageWidth := 595.28
	pageHeight := 841.89

	textWidth := pageWidth - leftMargin - rightMargin
	textHeight := pageHeight - topMargin - bottomMargin

	pdf.SetXY(leftMargin, topMargin)
	rect := gopdf.Rect{W: textWidth, H: textHeight}
	pdf.MultiCell(&rect, doc.Content)

	unsignedFilename := "unsigned_" + filename
	if err := pdf.WritePdf(unsignedFilename); err != nil {
		return fmt.Errorf("failed to write unsigned PDF: %w", err)
	}

	pfxData, err := os.ReadFile("cert.p12")
	if err != nil {
		return fmt.Errorf("failed to read cert.p12: %w", err)
	}

	privateKey, certificate, err := pkcs12.Decode(pfxData, "123456")
	if err != nil {
		return fmt.Errorf("failed to decode pfx: %w", err)
	}

	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("private key is not RSA")
	}

	signData := sign.SignData{
		Certificate:     certificate,
		Signer:          rsaKey,
		DigestAlgorithm: crypto.SHA256,
	}

	fmt.Println("Signing file:", unsignedFilename, " -> ", filename)
	if err := sign.SignFile(unsignedFilename, filename, signData); err != nil {
		fmt.Println("SignFile error:", err)
		return fmt.Errorf("PDF signing failed: %w", err)
	}

	_ = os.Remove(unsignedFilename)

	return nil
}
