package ansigo

func NewTagMatcher(start byte, mid []byte, end byte, unknownLength bool) *tagMatcher {
	t := tagMatcher{
		startByte:  start,
		midBytes:   mid,
		endByte:    end,
		exactMatch: !unknownLength,
		totalSize:  uint8(len(mid) + 1), // total size without the end byte
		position:   0,
	}
	return &t
}

type tagMatcher struct {
	exactMatch bool // lets unknown bytes keep matching before the end character is found
	totalSize  uint8
	position   uint8
	startByte  byte
	endByte    byte
	midBytes   []byte
}

func (t *tagMatcher) Seek(pos uint8) {
	t.position = pos
}

func (t *tagMatcher) MatchNext(char byte) (matched bool, complete bool) {

	if t.position == 0 {

		// Look for starting byte
		if char == t.startByte {
			t.position++
			// Still matching
			return true, false
		}
		// Failed to match.
		t.position = 0
		return false, true
	}

	if t.position >= t.totalSize {

		// Look for ending byte
		if char == t.endByte {
			t.position++
			// Matched and done.
			return true, true
		}

		if t.exactMatch {
			// Failed to match and required exact
			t.position = 0
			return false, true
		}

		// allows more characters before finishing.
		return true, false
	}

	// Look for mid bytes match
	if char == t.midBytes[t.position-1] {
		t.position++
		// Still matching
		return true, false
	}

	// Failed to match
	t.position = 0
	return false, true
}

func (t *tagMatcher) Reset() {
	t.Seek(0)
}
