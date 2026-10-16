package virtualcamera

type VirtualCamera interface {
	Start() error
	WriteFrame(frame []byte) error
	Stop() error
}
