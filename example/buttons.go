// Program buttons polls for button presses.
package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gonzojive/sane"
	"golang.org/x/image/tiff"
)

func doWork() error {
	if err := sane.Init(); err != nil {
		return fmt.Errorf("error initializing sane library: %v", err)
	}
	defer sane.Exit()

	devs, err := sane.Devices()
	if err != nil {
		return err
	}
	if len(devs) == 0 {
		return fmt.Errorf("SANE did not return any scanner devices.")
	}
	log.Printf("SANE returned %d scanner devices:", len(devs))
	for _, dev := range devs {
		log.Printf("  %s", dev.Name)
	}

	c, err := sane.Open(devs[0].Name)
	if err != nil {
		return fmt.Errorf("could not Open(%q): %v", devs[0].Name, err)
	}
	defer c.Close()

	log.Printf("%d options available for scanner:", len(c.Options()))
	for _, opt := range c.Options() {
		if strings.Contains(strings.ToLower(opt.Name+opt.Desc+opt.Title+opt.Group), "button") || true {
			log.Printf("  - %q: [type %q] %v+", opt.Name, opt.Type, opt)
		}
	}
	return monitorButtons(c)
}

func monitorButtons(c *sane.Conn) error {
	buttons := []string{"file", "extra", "scan", "copy", "email"} // Option names for buttons.
	values := make(map[string]bool)
	for _, b := range buttons {
		values[b] = false
	}
	for i := 0; i < 1000000; i++ {
		for _, button := range buttons {
			val, err := c.GetOption(button)
			if err != nil {
				return fmt.Errorf("GetOption(%q) failed: %v", button, err)
			}
			valB, ok := val.(bool)
			if !ok {
				return fmt.Errorf("expected boolean, got %v", val)
			}
			if valB != values[button] {
				values[button] = valB
				log.Printf("GetOption(%q) = %v", button, val)
			}
		}
		time.Sleep(time.Millisecond * 20)
		if values["scan"] {
			if _, err := c.SetOption("resolution", 1200); err != nil {
				return fmt.Errorf("error setting resolution: %v", err)
			}
			if _, err := c.SetOption("mode", "Color"); err != nil {
				return fmt.Errorf("error setting mode: %v", err)
			}
			doScan(c, "scan.png")
		}
		if values["file"] {
			if _, err := c.SetOption("resolution", 300); err != nil {
				return fmt.Errorf("error setting resolution: %v", err)
			}
			if _, err := c.SetOption("mode", "Color"); err != nil {
				return fmt.Errorf("error setting mode: %v", err)
			}
			doScan(c, "scan.png")
		}
	}
	return nil
}

func main() {
	if err := doWork(); err != nil {
		log.Printf("Fatal error: %v", err)
		os.Exit(1)
	}
}

func doScan(c *sane.Conn, fileName string) error {
	_, err := pathToEncoder(fileName)
	if err != nil {
		return err
	}

	if err := c.Start(); err != nil {
		return err
	}
	defer c.Cancel()

	p, err := c.Params()
	if err != nil {
		return err
	}

	log.Printf("scanning image with params %v", p)

	buf := make([]byte, 64*1024)
	numBytes := 0
	for {
		bytesRead, err := c.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		numBytes += bytesRead
	}

	log.Printf("Scanned image with %d bytes", numBytes)
	return nil

	// img, err := c.ReadImage()
	// if err != nil {
	// 	return err
	// }
	// log.Printf("Scanned image with bounds %v", img.Bounds())
	// return nil

	/*
		f, err := os.Create(fileName)
		if err != nil {
			die(err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				die(err)
			}
		}()

		c, err := openDevice(deviceName)
		if err != nil {
			die(err)
		}
		defer c.Close()
	*/

	/*
		if err := enc(f, img); err != nil {
			return err
		}
	*/
}

type EncodeFunc func(io.Writer, image.Image) error

func pathToEncoder(path string) (EncodeFunc, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return png.Encode, nil
	case ".jpg", ".jpeg":
		return func(w io.Writer, m image.Image) error {
			return jpeg.Encode(w, m, nil)
		}, nil
	case ".tif", ".tiff":
		return func(w io.Writer, m image.Image) error {
			return tiff.Encode(w, m, nil)
		}, nil
	default:
		return nil, fmt.Errorf("unrecognized extension")
	}
}
