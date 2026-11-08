-- +goose Up
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE faces (
    id BIGSERIAL PRIMARY KEY,
    image_name VARCHAR(255) NOT NULL,
    embedding VECTOR(512),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX faces_embedding_idx
ON faces
USING hnsw (embedding vector_cosine_ops);

-- +goose Down
SELECT 'down SQL query';
