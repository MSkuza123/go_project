DROP TABLE IF EXISTS pdf_documents;

CREATE ROLE pdf_documents WITH LOGIN PASSWORD 'pa55w0rd';

CREATE TABLE pdf_documents (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    pdf BYTEA NOT NULL
);

GRANT SELECT, INSERT, UPDATE, DELETE ON pdf_documents TO pdf_documents;
GRANT USAGE, SELECT ON SEQUENCE pdf_documents_id_seq TO pdf_documents;
