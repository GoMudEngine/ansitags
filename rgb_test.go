package ansitags

import (
	"reflect"
	"testing"
)

func TestNewRGB(t *testing.T) {
	tests := []struct {
		name    string
		r, g, b uint8
		wantHex string
	}{
		{"black", 0, 0, 0, "000000"},
		{"white", 255, 255, 255, "ffffff"},
		{"purple", 128, 0, 128, "800080"},
		{"mid-gray", 16, 32, 48, "102030"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := newRGB(tc.r, tc.g, tc.b)
			if got.R != tc.r || got.G != tc.g || got.B != tc.b {
				t.Errorf("newRGB(%d,%d,%d) = rgb{%d,%d,%d,...}; want rgb{%d,%d,%d,...}",
					tc.r, tc.g, tc.b, got.R, got.G, got.B, tc.r, tc.g, tc.b)
			}
			if got.Hex != tc.wantHex {
				t.Errorf("newRGB(%d,%d,%d).Hex = %q; want %q", tc.r, tc.g, tc.b, got.Hex, tc.wantHex)
			}
		})
	}
}

func TestRGB_Lookups(t *testing.T) {

	var tests = []struct {
		name string
		code int
		want rgb
	}{
		// out of range clamps to black
		{"below-range", -1, rgb{0, 0, 0, "000000"}},
		{"above-range", 256, rgb{0, 0, 0, "000000"}},

		// standard palette (0–15)
		{"code 0", 0, rgb{0, 0, 0, "000000"}},
		{"code 1", 1, rgb{128, 0, 0, "800000"}},
		{"code 7", 7, rgb{192, 192, 192, "c0c0c0"}},
		{"code 15", 15, rgb{255, 255, 255, "ffffff"}},

		// first entries of 6×6×6 cube (16–231)
		{"cube 16", 16, rgb{0, 0, 0, "000000"}},
		{"cube 17", 17, rgb{0, 0, 95, "00005f"}},
		{"cube 231", 231, rgb{255, 255, 255, "ffffff"}},

		// grayscale ramp (232–255)
		{"gray 232", 232, rgb{8, 8, 8, "080808"}},
		{"gray 255", 255, rgb{238, 238, 238, "eeeeee"}},

		// magenta test
		{"magenta 201", 201, rgb{255, 0, 255, "ff00ff"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := RGB(tc.code)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("RGB(%d) = %+v; want %+v", tc.code, got, tc.want)
			}
		})
	}
}
