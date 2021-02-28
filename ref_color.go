package amongus

import (
	"image/color"
)

type ColorRef Ref

func (r ColorRef) Read() color.Color {
	var data [4]float32
	Ref(r).Read(&data)
	return color.RGBA{
		uint8(data[0] * 255),
		uint8(data[1] * 255),
		uint8(data[2] * 255),
		uint8(data[3] * 255),
	}
}

func (r ColorRef) Write(c color.Color) {
	re, g, b, a := c.RGBA()
	data := [4]float32{
		float32(re) / 0xffff,
		float32(g) / 0xffff,
		float32(b) / 0xffff,
		float32(a) / 0xffff,
	}
	Ref(r).Write(&data)
}
