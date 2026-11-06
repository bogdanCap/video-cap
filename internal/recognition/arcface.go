package recognition

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"sync"

	ort "github.com/yalue/onnxruntime_go"
	"golang.org/x/image/draw"
)

const (
	faceSize      = 112
	embeddingSize = 512
)

type ArcFace struct {
	mu sync.Mutex

	session      *ort.AdvancedSession
	inputTensor  *ort.Tensor[float32]
	outputTensor *ort.Tensor[float32]
}

// NewArcFace initializes the ArcFace ONNX model.
func NewArcFace(
	modelPath string,
	inputName string,
	outputName string,
) (*ArcFace, error) {

	// Create input tensor: [1, 112, 112, 3].
	inputTensor, err := ort.NewTensor(
		ort.NewShape(1, faceSize, faceSize, 3),
		make([]float32, faceSize*faceSize*3),
	)
	if err != nil {
		return nil, err
	}

	// Create output tensor: [1, 512].
	outputTensor, err := ort.NewEmptyTensor[float32](
		ort.NewShape(1, embeddingSize),
	)
	if err != nil {
		inputTensor.Destroy()
		return nil, err
	}

	// Load ONNX model.
	session, err := ort.NewAdvancedSession(
		modelPath,
		[]string{inputName},
		[]string{outputName},
		[]ort.Value{inputTensor},
		[]ort.Value{outputTensor},
		nil,
	)
	if err != nil {
		inputTensor.Destroy()
		outputTensor.Destroy()
		return nil, err
	}

	return &ArcFace{
		session:      session,
		inputTensor:  inputTensor,
		outputTensor: outputTensor,
	}, nil
}

// Embed generates a normalized face embedding.
func (a *ArcFace) Embed(
	ctx context.Context,
	img image.Image,
) ([]float32, error) {

	if img == nil {
		return nil, errors.New("image is nil")
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// 1. Resize image to 112x112.
	resized := image.NewRGBA(
		image.Rect(0, 0, faceSize, faceSize),
	)

	draw.CatmullRom.Scale(
		resized,
		resized.Bounds(),
		img,
		img.Bounds(),
		draw.Over,
		nil,
	)

	// 2. Convert RGB pixels into normalized float32 values.
	inputData := a.inputTensor.GetData()

	for y := 0; y < faceSize; y++ {
		for x := 0; x < faceSize; x++ {

			pixel := color.NRGBAModel.Convert(
				resized.At(x, y),
			).(color.NRGBA)

			index := (y*faceSize + x) * 3

			inputData[index] =
				(float32(pixel.R) - 127.5) / 128.0

			inputData[index+1] =
				(float32(pixel.G) - 127.5) / 128.0

			inputData[index+2] =
				(float32(pixel.B) - 127.5) / 128.0
		}
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// 3. Execute ArcFace neural network.
	if err := a.session.Run(); err != nil {
		return nil, fmt.Errorf("arcface inference failed: %w", err)
	}

	// 4. Extract embedding.
	outputData := a.outputTensor.GetData()

	if len(outputData) != embeddingSize {
		return nil, fmt.Errorf(
			"invalid embedding size: %d",
			len(outputData),
		)
	}

	embedding := make([]float32, embeddingSize)

	var norm float64

	for i, value := range outputData {

		if math.IsNaN(float64(value)) ||
			math.IsInf(float64(value), 0) {
			return nil, errors.New("invalid embedding value")
		}

		embedding[i] = value
		norm += float64(value * value)
	}

	// 5. L2 normalization.
	norm = math.Sqrt(norm)

	if norm == 0 {
		return nil, errors.New("zero embedding norm")
	}

	for i := range embedding {
		embedding[i] /= float32(norm)
	}

	return embedding, nil
}

// Close releases ONNX Runtime resources.
func (a *ArcFace) Close() error {

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.session != nil {
		if err := a.session.Destroy(); err != nil {
			return err
		}
		a.session = nil
	}

	if a.inputTensor != nil {
		if err := a.inputTensor.Destroy(); err != nil {
			return err
		}
		a.inputTensor = nil
	}

	if a.outputTensor != nil {
		if err := a.outputTensor.Destroy(); err != nil {
			return err
		}
		a.outputTensor = nil
	}

	return nil
}
