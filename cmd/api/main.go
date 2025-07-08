package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"pdfcrudsign/internal/data"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type config struct {
	port int
	env  string
	dsn  string
}

type application struct {
	config config
	logger *log.Logger
	models data.Models
}

func main() {

	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "dev", "Environment (dev|staging|prod)")
	flag.StringVar(&cfg.dsn, "db-dsn", os.Getenv("DOCUMENT_PDF_DB_DSN"), "PostgreSQL DSN")
	flag.Parse()

	if cfg.dsn == "" {
		cfg.dsn = "postgres://postgres:admin@localhost/pdf_documents?sslmode=disable"
		// cfg.dsn = "postgres://postgres:admin@localhost:5433/pdf_documents?sslmode=disable"
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := sql.Open("postgres", cfg.dsn)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		logger.Fatal(err)
	}

	logger.Printf("database connection pool established")

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	addr := fmt.Sprintf(":%d", cfg.port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Printf("starting %s server on %s", cfg.env, addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

func (app *application) routes() http.Handler {
	// router := gin.Default()
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	gin.SetMode(gin.DebugMode)

	router.POST("/pdfs", app.insertPDFHandler)
	router.GET("/pdfs", app.getAllPDFsHandler)
	router.GET("/pdfs/:id", app.getPDFHandler)
	router.PATCH("/pdfs/:id", app.updatePDFHandler)
	router.DELETE("/pdfs/:id", app.deletePDFHandler)

	return router
}
