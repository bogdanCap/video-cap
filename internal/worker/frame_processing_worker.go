package worker

import (
	"context"
	"log"
	"time"
	"sync"

	"github.com/bogdanCap/video-cap/internal/preview"
	"github.com/bogdanCap/video-cap/internal/recording"
	"github.com/bogdanCap/video-cap/internal/facemove"
	//"github.com/bogdanCap/video-cap/internal/detection"
)

const previewTimeout = 100 * time.Millisecond
const recordingTimeout = 40 * time.Millisecond


type FrameProcessingWorker struct {
	preview  *preview.FynePreview
	recorder recording.Recorder
	faceMotion *facemove.MotionTracker

	previewChan   chan []byte
	recordingChan chan []byte
	//faceMotionChan chan detection.Face

	wg sync.WaitGroup
	//protect from - two goroutines could write to the same FFmpeg stdin at the same time.
	recordingMu sync.Mutex
}

func NewFrameProcessingWorker(
	videoPreview *preview.FynePreview,
	recorder recording.Recorder,
	faceMotion *facemove.MotionTracker,
) *FrameProcessingWorker {
	return &FrameProcessingWorker{
		preview:  videoPreview,
		recorder: recorder,
		faceMotion: faceMotion,

		// Small buffer because preview can drop frames.
		previewChan: make(chan []byte, 1),

		// Larger buffer because recording must not drop frames.
		recordingChan: make(chan []byte, 30),

		//faceMotionChan: make(chan detection.Face, 1),

	}
}

func (w *FrameProcessingWorker) Run(ctx context.Context/*, chanFaceMotionResultChan chan<- facemove.MotionResult*/) {
	//listening channel i run logic
	/**TODO wg.ADD replace with (and defer do not need) + tested how its works
	 wg.Go(func() {}

	 workersList := []func(context.Context){
		worker.ProcessData, // Element 1
		worker.LogAction,   // Element 2
	}

	// Loop through the array and call each method
	for _, worker := range workersList {
		wg.Go(func() {
			worker(ctx)
		})
	}
	**/
	
	w.wg.Add(2)
	
	go func() {
		defer w.wg.Done()

		w.previewWorker(ctx)
	}()


	go func() {
		defer w.wg.Done()

		w.recordingWorker(ctx)
	}()

	/* face motion
	go func() {
		defer w.wg.Done()

		w.faceMotionWorker(ctx, chanFaceMotionResultChan)
	}()*/
}

func (w *FrameProcessingWorker) Wait() {
	w.wg.Wait()
}

func (w *FrameProcessingWorker) ProduceFrames(
	ctx context.Context,
	frameChan <-chan []byte,
	//frame []byte,
	//faceFrame detection.Face,
) {
	for {
		select {
		case <-ctx.Done():
			log.Println("produce frame stopped")

			return

		case frame, ok := <-frameChan:
			if !ok {
				return
			}

			select {
			case w.previewChan <- frame:
			default:
				// Drop preview frame.
			}

			select {
			case w.recordingChan <- frame:
			case <-ctx.Done():
				return
			}
		}
	}
	/*
	// send/push job data to the worker
	//this select need to detect cancel context from preview and recording goroutines
	// Preview can drop frames if it is behind.
	select {
	case w.previewChan <- frame:
	default:
		// Drop preview frame - if preview is behind recording goroutines
	}

	// Recording must not drop frames.
	// If the buffer is full, this blocks until
	// the recording worker consumes a frame.
	select {
	case w.recordingChan <- frame:
	case <-ctx.Done():
		return
	}
		*/

	//face motion detection
	/*
	select {
	case w.faceMotionChan <- faceFrame:
	case <-ctx.Done():
		return
	}*/
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

/*
func (w *FrameProcessingWorker) faceMotionWorker(
	ctx context.Context,
	chanFaceMotionResultChan chan<- facemove.MotionResult,
) {
	for {
		select {
		case <-ctx.Done():
			log.Println("face motion worker stopped")
			return

		case faceFrame, ok := <-w.faceMotionChan:
			if !ok {
				return
			}

			type trackingResponse struct {
				isMoving bool
				face     detection.Face
				err      error
			}
			done := make(chan trackingResponse, 1)

			//done := make(chan bool, 1)

			// Run ShowFrame in another goroutine - this needed for timeout - if ShowFrame freez - code continue to work.
			go func() {
				isMoving, matchedFace, err := w.faceMotion.TrackMovement(faceFrame)

				log.Println("face is move:  ", isMoving)

				if err != nil {
				//	continue
					log.Println("error to track face motion:  ", err)
				}

				done <- trackingResponse{
					isMoving: isMoving,
					face:     matchedFace,
					err:      err,
				}
				
			}()


			timer := time.NewTimer(previewTimeout)

			select {
			case out := <-done:
				timer.Stop()

				if out.err != nil {
					log.Println("preview:", out.err)
					continue
				}
				
				// ✅ FIXED: Pure non-blocking send. No read operations on send-only channel.
				select {
				case chanFaceMotionResultChan <- facemove.MotionResult{
					IsMoving: out.isMoving,
					Face:     out.face,
				}:
					log.Println("event caught before send event to side button")
				
				default:
					// If the "Face detect" button loop hasn't started yet or is full,
					// it hits this default case and drops the frame gracefully so the 
					// main camera/recording loop never freezes.
					log.Println("chanFaceMotionResultChan is full or UI loop is inactive - frame dropped")
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
*/
func (w *FrameProcessingWorker) writeFrame(frame []byte) error {
	// Protect FFmpeg stdin from concurrent WriteFrame calls.
	// protect from - two goroutines could write to the same FFmpeg stdin at the same time.
	w.recordingMu.Lock()
	defer w.recordingMu.Unlock()

	return w.recorder.WriteFrame(frame)
}
