package camera

import (
    "fmt"
)

type CameraType int

const (
	USBType CameraType = iota + 1 // 1
	CSIType                       // 2
	CSIDualType                   // 3 (in this case csi camera record and preview pipline going on rpicam-vid level)
)

// 3. Implement fmt.Stringer for readable string output
func (i CameraType) String() string {
	switch i {
	case USBType:
		return "usb"
	case CSIType:
		return "sci"
	case CSIDualType:
		return "sci-dual"
	default:
		return "unknown"
	}
}

func (t CameraType) IsValid() bool {
	return t == USBType || t == CSIType || t == CSIDualType
}

// ParseInterfaceType safely converts an int from config into InterfaceType
func CameraTypeFromInt(v int) (CameraType, error) {
	cameraType := CameraType(v)

	if !cameraType.IsValid() {
		return 0, fmt.Errorf("unknown interface type: %d", v)
	}

	return cameraType, nil
}
