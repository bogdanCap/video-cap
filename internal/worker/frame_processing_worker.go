package worker

import (
	"context"
	"log"
	"time"

	"video/internal/preview"
	"video/internal/recording"
)

const previewTimeout = 100 * time.Millisecond
const recordingTimeout = 40 * time.Millisecond


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
}

func (w *FrameProcessingWorker) previewWorker(
	ctx context.Context,
) {
	for {
		select {
		case <-ctx.Done():
			log.Println("preview worker stopped")
			return

		case frame, ok := <-w.previewChan:
			if !ok {
				return
			}

			done := make(chan error, 1)

			// Run ShowFrame in another goroutine - this needed for timeout - if ShowFrame freez - code continue to work.
			go func() {
				done <- w.preview.ShowFrame(frame)
			}()

			timer := time.NewTimer(previewTimeout)

			select {
			case err := <-done:
				timer.Stop()

				if err != nil {
					log.Println("preview:", err)
				}

			case <-timer.C:
				log.Println("preview timeout - skip current job")

				// Do not wait for ShowFrame().
				// Continue the for loop and receive the next frame.

			case <-ctx.Done():
				timer.Stop()
				return
			}
			/*
			if err := w.preview.ShowFrame(frame); err != nil {
				log.Println("preview:", err)
			}*/
		}
	}
}

func (w *FrameProcessingWorker) recordingWorker(
	ctx context.Context,
) {
	/*
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
	}*/
	
	for {
		select {
		case <-ctx.Done():
			log.Println("stopping recording worker...")

			// Process all frames that are already
			// waiting in recordingChan.
			for {
				select {
				case frame := <-w.recordingChan:
					if err := w.recorder.WriteFrame(frame); err != nil {
						log.Println("recording:", err)
					}

				default:
					log.Println("recording worker stopped")
					return
				}
			}

		//case frame := <-w.recordingChan:
		//	if err := w.recorder.WriteFrame(jobCtx, frame); err != nil {
		//		log.Println("recording:", err)
		//	}
		//}
		case frame, ok := <-w.recordingChan:
			if !ok {
				return
			}

			done := make(chan error, 1)

			// Run WriteFrame in another goroutine, this needed for timeout - if ShowFrame freez - code continue to work.
			go func() {
				done <- w.recorder.WriteFrame(frame)
			}()

			timer := time.NewTimer(recordingTimeout)

			select {
			case err := <-done:
				timer.Stop()

				if err != nil {
					log.Println("recording:", err)
				}

			case <-timer.C:
				log.Println("recording timeout - skip current job")

				// Do not wait for WriteFrame().
				// Continue the for loop and receive the next frame.

			case <-ctx.Done():
				timer.Stop()
				return
			}
		}
	}
}
