package recording

type Recorder interface {
	Start() error
	WriteFrame(frame []byte) error
	Stop() error
}
