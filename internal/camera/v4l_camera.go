package camera

import (
	"fmt"

	"github.com/korandiz/v4l"
	"github.com/korandiz/v4l/fmt/mjpeg"
)

type V4LCamera struct {
	device *v4l.Device
}

func NewV4LCamera(
	devicePath string,
	width int,
	height int,
	fps uint32,
) (*V4LCamera, error) {
	device, err := v4l.Open(devicePath)
	if err != nil {
		return nil, fmt.Errorf("open camera: %w", err)
	}

	config := v4l.DeviceConfig{
		Format: mjpeg.FourCC,
		Width:  width,
		Height: height,
		FPS: v4l.Frac{
			N: 1,
			D: uint32(fps),
		},
	}

	if err := device.SetConfig(config); err != nil {
		device.Close()

		return nil, fmt.Errorf(
			"set camera config: %w",
			err,
		)
	}

	actualConfig, err := device.GetConfig()
	if err != nil {
		device.Close()

		return nil, fmt.Errorf(
			"get camera config: %w",
			err,
		)
	}

	fmt.Printf(
		"Camera config: format=%v %dx%d @ %d/%d FPS\n",
		actualConfig.Format,
		actualConfig.Width,
		actualConfig.Height,
		actualConfig.FPS.N,
		actualConfig.FPS.D,
	)

	//enabled camera device
	if err := device.TurnOn(); err != nil {
		device.Close()

		return nil, fmt.Errorf(
			"turn on camera: %w",
			err,
		)
	}

	return &V4LCamera{
		device: device,
	}, nil
	/*
	device, err := v4l.Open(devicePath)
	if err != nil {
		return nil, fmt.Errorf("open camera: %w", err)
	}

	configs, err := device.ListConfigs()
	if err != nil {
		device.Close()

		return nil, fmt.Errorf("list camera configs: %w", err)
	}

	var selected *v4l.DeviceConfig

	for _, cfg := range configs {

		fmt.Printf(
			"Camera config: format=%08x %dx%d fps=%d/%d\n",
			cfg.Format,
			cfg.Width,
			cfg.Height,
			cfg.FPS.N,
			cfg.FPS.D,
		)

		if cfg.Width == width &&
			cfg.Height == height &&
			cfg.FPS.N == fps &&
			cfg.FPS.D == 1 {

			cfgCopy := cfg
			selected = &cfgCopy

			break
		}
	}

	if selected == nil {
		device.Close()

		return nil, fmt.Errorf(
			"camera configuration %dx%d @ %d FPS not found",
			width,
			height,
			fps,
		)
	}

	if err := device.SetConfig(*selected); err != nil {
		device.Close()

		return nil, fmt.Errorf(
			"configure camera: %w",
			err,
		)
	}

	fmt.Printf(
		"Camera configured: %dx%d @ %d/%d FPS\n",
		selected.Width,
		selected.Height,
		selected.FPS.N,
		selected.FPS.D,
	)

	return &V4LCamera{
		device: device,
	}, nil*/
}
/*
func (c *V4LCamera) Start() error {
	if err := c.device.TurnOn(); err != nil {
		return fmt.Errorf("start camera: %w", err)
	}

	return nil
}*/

func (c *V4LCamera) Read() ([]byte, error) {
	buf, err := c.device.Capture()
	if err != nil {
		return nil, fmt.Errorf("capture frame: %w", err)
	}

	// Copy the frame from the V4L2 buffer.
	frame := make([]byte, buf.Len())

	n, err := buf.Read(frame)
	if err != nil {
		return nil, fmt.Errorf("read frame: %w", err)
	}

	return frame[:n], nil
}

func (c *V4LCamera) Stop() error {
	c.device.TurnOff()
	//if err := c.device.TurnOff(); err != nil {
	//	return fmt.Errorf("stop camera: %w", err)
	//}

	return nil
}

func (c *V4LCamera) Close() error {
	c.device.Close()
	
	return nil
}
