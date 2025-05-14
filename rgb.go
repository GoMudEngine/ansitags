package ansitags

// Accepts a 4-bit or 8-bit ANSI color code and returns an rgb struct for it.
// Usage:
//
// c := ansitags.RGB(201) // Magenta AKA #FF00FF
// fmt.Println("R:", c.R, "G:", c.G, "B:", c.B)
// fmt.Println("Hex:", c.Hex)
//
// Output:
// R: 255 G: 0 B: 255
// Hex: ff00ff
var (
	ansi256 [256]rgb
)

// rgb holds 24-bit color components and its hexadecimal representation.
type rgb struct {
	R, G, B uint8
	Hex     string
}

// RGB returns an rgb struct for codes 0–255, or black for out-of-range.
func RGB(colorCode int) rgb {
	if colorCode < 0 || colorCode > 255 {
		return newRGB(0, 0, 0)
	}
	return ansi256[colorCode]
}

// newRGB constructs an rgb and pre‐computes its hex string.
func newRGB(r, g, b uint8) rgb {
	const hexdigits = "0123456789abcdef"
	var buf [6]byte
	buf[0] = hexdigits[r>>4]
	buf[1] = hexdigits[r&0xF]
	buf[2] = hexdigits[g>>4]
	buf[3] = hexdigits[g&0xF]
	buf[4] = hexdigits[b>>4]
	buf[5] = hexdigits[b&0xF]
	return rgb{R: r, G: g, B: b, Hex: string(buf[:])}
}

// Pre-compute a look up table to avoid repeat calculations for only 256 possible values
func init() {
	// 0–15: standard + high-intensity
	base := [16]rgb{
		newRGB(0, 0, 0),
		newRGB(128, 0, 0),
		newRGB(0, 128, 0),
		newRGB(128, 128, 0),
		newRGB(0, 0, 128),
		newRGB(128, 0, 128),
		newRGB(0, 128, 128),
		newRGB(192, 192, 192),
		newRGB(128, 128, 128),
		newRGB(255, 0, 0),
		newRGB(0, 255, 0),
		newRGB(255, 255, 0),
		newRGB(0, 0, 255),
		newRGB(255, 0, 255),
		newRGB(0, 255, 255),
		newRGB(255, 255, 255),
	}
	copy(ansi256[:16], base[:])

	// 6×6×6 color cube: values = [0,95,135,175,215,255]
	cube := [6]uint8{0, 95, 135, 175, 215, 255}
	idx := 16
	for r := 0; r < 6; r++ {
		for g := 0; g < 6; g++ {
			for b := 0; b < 6; b++ {
				ansi256[idx] = newRGB(cube[r], cube[g], cube[b])
				idx++
			}
		}
	}

	// 232–255: grayscale ramp (8 + 10×step)
	for i := 232; i < 256; i++ {
		gray := uint8(8 + (i-232)*10)
		ansi256[i] = newRGB(gray, gray, gray)
	}
}
