package detection

import (
    "errors"
    "image"
)

func CropFace(faceImage FaceImage) (image.Image, error) {
    if faceImage.Image == nil {
        return nil, errors.New("image is nil")
    }

    face := faceImage.Face
    bounds := faceImage.Image.Bounds()

    rect := image.Rect(
        face.X,
        face.Y,
        face.X+face.Width,
        face.Y+face.Height,
    )

    rect = rect.Intersect(bounds)

    if rect.Empty() {
        return nil, errors.New("invalid face coordinates")
    }

    cropped := image.NewRGBA(image.Rect(
        0,
        0,
        rect.Dx(),
        rect.Dy(),
    ))

    for y := 0; y < rect.Dy(); y++ {
        for x := 0; x < rect.Dx(); x++ {
            cropped.Set(x, y, faceImage.Image.At(
                rect.Min.X+x,
                rect.Min.Y+y,
            ))
        }
    }

    return cropped, nil
}
