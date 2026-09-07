package worker

import (
	"context"
	"log"

	"video/internal/preview"
	"video/internal/recording"
)

type FrameProcessingWorker struct {
	preview  *preview.FynePreview
	recorder recording.Recorder

	previewChan   chan []byte
	recordingChan chan []byte
}

func NewFrameProcessingWorker(
	videoPreview *preview.FynePreview,
	recorder recording.Recorder,
) *FrameProcessingWorker {
	return &FrameProcessingWorker{
		preview:  videoPreview,
		recorder: recorder,

		// Small buffer because preview can drop frames.
		previewChan: make(chan []byte, 2),

		// Larger buffer because recording must not drop frames.
		recordingChan: make(chan []byte, 30),
	}
}

func (w *FrameProcessingWorker) Run(ctx context.Context) {
	go func() {
		w.previewWorker(ctx)
	}()


	go func() {
		w.recordingWorker(ctx)
	}()
}

func (w *FrameProcessingWorker) Process(
	ctx context.Context,
	frame []byte,
) {
	// Preview can drop frames if it is behind.
	select {
	case w.previewChan <- frame:
	default:
		// Drop preview frame.
	}

	// Recording must not drop frames.
	// If the buffer is full, this blocks until
	// the recording worker consumes a frame.
	select {
	case w.recordingChan <- frame:
	case <-ctx.Done():
		return
	}
	//w.recordingChan <- frame
}

func (w *FrameProcessingWorker) previewWorker(
	ctx context.Context,
) {
	for {
		select {
		case <-ctx.Done():
			log.Println("preview worker stopped")
			return

		case frame := <-w.previewChan:
			if err := w.preview.ShowFrame(frame); err != nil {
				log.Println("preview:", err)
			}
		}
	}
}

func (w *FrameProcessingWorker) recordingWorker(
	ctx context.Context,
) {
	for {
		select {
		case <-ctx.Done():
			log.Println("recording worker stopped")
			return

		case frame := <-w.recordingChan:
			if err := w.recorder.WriteFrame(frame); err != nil {
				log.Println("recording:", err)
			}
		}
	}
}
