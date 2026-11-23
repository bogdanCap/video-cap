package camera

import (
    "fmt"
)

type CameraType int

const (
	USBType CameraType = iota + 1 // 1
	CSIType                       // 2
)

// 3. Implement fmt.Stringer for readable string output
func (i CameraType) String() string {
	switch i {
	case USBType:
		return "usb"
	case CSIType:
		return "sci"
	default:
		return "unknown"
	}
}

func (i CameraType) IsValid() bool {
	return i == USBType || i == CSIType
}

// ParseInterfaceType safely converts an int from config into InterfaceType
func CameraTypeFromInt(v int) (CameraType, error) {
	cameraType := CameraType(v)

	if !cameraType.IsValid() {
		return 0, fmt.Errorf("unknown interface type: %d", v)
	}

	return cameraType, nil
}
