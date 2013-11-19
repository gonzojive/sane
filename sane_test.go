// Copyright (C) 2013 Tiago Quelhas. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sane

import (
	"image/color"
	"testing"
)

const TestDevice = "test"

var (
	monoBlack  = color.Gray{0}
	monoWhite  = color.Gray{255}
	colorBlack = color.RGBA{0, 0, 0, 255}
	colorWhite = color.RGBA{255, 255, 255, 255}
)

var typeMap = map[Type]string{
	TypeBool:   "bool",
	TypeInt:    "int",
	TypeFloat:  "float",
	TypeString: "string",
	TypeButton: "button",
}

var unitMap = map[Unit]string{
	UnitNone:    "none",
	UnitPixel:   "pixel",
	UnitBit:     "bit",
	UnitMm:      "mm",
	UnitDpi:     "dpi",
	UnitPercent: "percent",
	UnitMsec:    "milliseconds",
}

func runTest(t *testing.T, f func(c *Conn)) {
	if err := Init(); err != nil {
		t.Fatal("init failed:", err)
	}
	defer Exit()
	c, err := Open(TestDevice)
	if err != nil {
		t.Fatal("open failed:", err)
	}
	defer c.Close()
	f(c)
}

func setOption(t *testing.T, c *Conn, name string, val interface{}) Info {
	i, err := c.SetOption(name, val)
	if err != nil {
		t.Fatalf("set option %s to %v failed: %v", name, val, err)
	}
	return i
}

func readImage(t *testing.T, c *Conn) (*Image, int, int) {
	i, err := c.ReadImage()
	if err != nil {
		t.Fatal("read image failed:", err)
	}
	b := i.Bounds()
	if b.Min.X != 0 || b.Min.Y != 0 || b.Max.X <= b.Min.X || b.Max.Y <= b.Min.Y {
		t.Fatal("bad bounds:", b)
	}
	return i, b.Max.X, b.Max.Y
}

func TestDevices(t *testing.T) {
	if _, err := Devices(); err != nil {
		t.Fatal("list devices failed:", err)
	}
}

func checkOptionType(t *testing.T, o *Option, val interface{}) {
	typeName := typeMap[o.Type]
	switch val.(type) {
	case bool:
		if o.Type != TypeBool {
			t.Errorf("option %s has type bool, should be %s", o.Name, typeName)
		}
	case int:
		if o.Type != TypeInt {
			t.Errorf("options %s has type int, should be %s", o.Name, typeName)
		}
	case float64:
		if o.Type != TypeFloat {
			t.Errorf("option %s has type float, should be %s", o.Name, typeName)
		}
	case string:
		if o.Type != TypeString {
			t.Errorf("option %s has type string, should be %s", o.Name, typeName)
		}
	default:
		t.Errorf("option %s has unexpected type, should be %s", o.Name, typeName)
	}
}

func TestOptions(t *testing.T) {
	runTest(t, func(c *Conn) {
		for _, o := range c.Options() {
			if _, ok := typeMap[o.Type]; !ok {
				t.Errorf("unknown type %d for option %s", o.Type, o.Name)
			}
			if _, ok := unitMap[o.Unit]; !ok {
				t.Errorf("unknown unit %d for option %s", o.Unit, o.Name)
			}
			if !o.IsActive {
				continue
			}
			if o.Type == TypeButton {
				return
			}
			val, err := c.GetOption(o.Name)
			if err != nil {
				t.Errorf("get option %s failed: %v", o.Name, err)
			} else {
				checkOptionType(t, &o, val)
			}
		}
	})
}

func TestGrayImage(t *testing.T) {
	runTest(t, func(c *Conn) {
		i, xMax, yMax := readImage(t, c)
		if i.ColorModel() != color.GrayModel {
			t.Fatal("bad color model")
		}
		for x := 0; x < xMax; x++ {
			for y := 0; y < yMax; y++ {
				if i.At(x, y) != monoBlack {
					t.Fatalf("bad pixel at (%d,%d)", x, y)
				}
			}
		}
	})
}

func TestColorImage(t *testing.T) {
	runTest(t, func(c *Conn) {
		_ = setOption(t, c, "mode", "Color")
		i, xMax, yMax := readImage(t, c)
		if i.ColorModel() != color.RGBAModel {
			t.Fatal("bad color model")
		}
		for x := 0; x < xMax; x++ {
			for y := 0; y < yMax; y++ {
				if i.At(x, y) != colorBlack {
					t.Fatalf("bad pixel at (%d,%d)", x, y)
				}
			}
		}
	})
}