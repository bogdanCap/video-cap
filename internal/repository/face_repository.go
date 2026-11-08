package repository

import (
    "context"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/pgvector/pgvector-go"
)

type FaceMatch struct {
    ID         int64
    ImagePath  string
    Similarity float32
}

type FaceRepository struct {
    db *pgxpool.Pool
}

func NewFaceRepository(db *pgxpool.Pool) *FaceRepository {
    return &FaceRepository{
        db: db,
    }
}

func (r *FaceRepository) FindSimilar(
    ctx context.Context,
    embedding []float32,
    threshold float32,
) (*FaceMatch, error) {

    /*
    SELECT
        id,
        image_name,
        1 - (embedding <=> $1::vector) AS similarity
    FROM faces
    ORDER BY embedding <=> $1::vector
    LIMIT 10;
    */

    query := `
        SELECT
            id,
            image_path,
            1 - (embedding <=> $1) AS similarity
        FROM faces
        WHERE embedding <=> $1 <= $2
        ORDER BY embedding <=> $1
        LIMIT 1
    `

    var match FaceMatch

    err := r.db.QueryRow(
        ctx,
        query,
        pgvector.NewVector(embedding),
        1.0-threshold,
    ).Scan(
        &match.ID,
        &match.ImagePath,
        &match.Similarity,
    )

    if err != nil {
        // pgx.ErrNoRows means no matching face was found.
        if err.Error() == "no rows in result set" {
            return nil, nil
        }

        return nil, err
    }

    return &match, nil
}

func (r *FaceRepository) SaveEmbedding(
    ctx context.Context,
    imagePath string,
    embedding []float32,
) (int64, error) {

    query := `
        INSERT INTO faces (image_path, embedding)
        VALUES ($1, $2)
        RETURNING id
    `

    var id int64

    err := r.db.QueryRow(
        ctx,
        query,
        imagePath,
        pgvector.NewVector(embedding),
    ).Scan(&id)

    if err != nil {
        return 0, err
    }

    return id, nil
}
