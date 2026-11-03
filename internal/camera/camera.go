package camera

// Camera describes a camera capable of streaming frames.
type Camera interface {
	Read() ([]byte, error)
	Stop() error
	Close() error
}
