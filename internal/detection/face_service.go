package detection

import (
    "context"
    "errors"
    "image"

    //"github.com/bogdanCap/video-cap/internal/detection"
    "github.com/bogdanCap/video-cap/internal/recognition"
	 "github.com/bogdanCap/video-cap/internal/repository"
)

type FaceRepository interface {
    FindSimilar(
        ctx context.Context,
        embedding []float32,
        threshold float32,
    ) (*repository.FaceMatch, error)

    SaveEmbedding(
        ctx context.Context,
        imagePath string,
        embedding []float32,
    ) (int64, error)
}

type FaceService struct {
    detector   *FaceDetector
    recognizer recognition.FaceRecognizer
    repository FaceRepository
}

func NewFaceService(
    detector *FaceDetector,
    recognizer recognition.FaceRecognizer,
    repository FaceRepository,
) *FaceService {
    return &FaceService{
        detector:   detector,
        recognizer: recognizer,
        repository: repository,
    }
}

func (s *FaceService) FindOrCreate(
    ctx context.Context,
    img image.Image,
    imagePath string,
) (int64, error) {

    // 1. Detect face using Pigo.
    faces, err := s.detector.Detect(img)
    if err != nil {
        return 0, err
    }

    if len(faces) == 0 {
        return 0, errors.New("face not found")
    }
	//TODO this logic move to video_control_service.go
	// into line 181 in camera_worker.go

    // 2. Extract face from original image.
    faceImage := FaceImage{
        Image: img,
        Face:  faces[0],
    }

    cropped, err := CropFace(faceImage)
    if err != nil {
        return 0, err
    }

    // 3. Generate ArcFace embedding.
    embedding, err := s.recognizer.Embed(ctx, cropped)
    if err != nil {
        return 0, err
    }

    // 4. Search for an existing face.
    existing, err := s.repository.FindSimilar(
        ctx,
        embedding,
        0.85,
    )

    if err != nil {
        return 0, err
    }

    // 5. If face already exists, return its ID.
    if existing != nil {
        return existing.ID, nil
    }

    // 6. Otherwise, save new face.
    id, err := s.repository.SaveEmbedding(
        ctx,
        imagePath,
        embedding,
    )

    if err != nil {
        return 0, err
    }

    return id, nil
}
