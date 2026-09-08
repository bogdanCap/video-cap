package worker

import (
	"context"
	"log"
	"time"
	"sync"

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

	wg sync.WaitGroup
	//protect from - two goroutines could write to the same FFmpeg stdin at the same time.
	recordingMu sync.Mutex
}

func NewFrameProcessingWorker(
	videoPreview *preview.FynePreview,
	recorder recording.Recorder,
) *FrameProcessingWorker {
	return &FrameProcessingWorker{
		preview:  videoPreview,
		recorder: recorder,

		// Small buffer because preview can drop frames.
		previewChan: make(chan []byte, 1),

		// Larger buffer because recording must not drop frames.
		recordingChan: make(chan []byte, 30),

	}
}

func (w *FrameProcessingWorker) Run(ctx context.Context) {
	w.wg.Add(2)
	
	go func() {
		defer w.wg.Done()

		w.previewWorker(ctx)
	}()


	go func() {
		defer w.wg.Done()

		w.recordingWorker(ctx)
	}()
}

func (w *FrameProcessingWorker) Wait() {
	w.wg.Wait()
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

				log.Println("preview worker stopped")
				return
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
			//stop button (cancel) logic
			// Process all frames that are already
			// waiting in recordingChan.
			for {
				select {
				case frame := <-w.recordingChan:
					if err := w.writeFrame(frame); err != nil {
						log.Println("recording:", err)
					}

				default:
					log.Println("recording worker stopped")
					return
				}
			}

		case frame, ok := <-w.recordingChan:
			if !ok {
				return
			}

			done := make(chan error, 1)

			// Run WriteFrame in another goroutine, this needed for timeout - if ShowFrame freez - code continue to work.
			go func() {
				//done <- w.recorder.WriteFrame(frame)
				done <- w.writeFrame(frame)
			}()

			timer := time.NewTimer(recordingTimeout)

			select {
			case err := <-done:
				//frame save and its ok
				timer.Stop()

				if err != nil {
					log.Println("recording:", err)
				}

			case <-timer.C:
				log.Println("recording timeout - skip current job")

				// Do not wait for WriteFrame().
				// Continue the for loop and receive the next frame.

			case <-ctx.Done():
				//Context cancellation -> main.go -> cancel()
				timer.Stop()
				return
			}
		}
	}
}

func (w *FrameProcessingWorker) writeFrame(frame []byte) error {
	// Protect FFmpeg stdin from concurrent WriteFrame calls.
	// protect from - two goroutines could write to the same FFmpeg stdin at the same time.
	w.recordingMu.Lock()
	defer w.recordingMu.Unlock()

	return w.recorder.WriteFrame(frame)
}
