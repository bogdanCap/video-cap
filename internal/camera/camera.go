package camera

// Camera describes a camera capable of streaming frames.
type Camera interface {
	Start() error
	Read() ([]byte, error)
	Stop() error
	Close() error
}
