# Building
1. Run docker container
   ```bash
   docker compose up -d

2. Set environmental variable
    ```bash
   $env:DOCUMENT_PDF_DB_DSN = "postgres://pdf_documents:pa55w0rd@localhost:5433/pdf_documents?sslmode=disable
3. Start application
    ```bash
   go run .cmd/api   
