package codec

import "testing"

func TestOption(t *testing.T) {
	opt := GetOption(GobCodec)
	if len(opt) != 5 {
		t.Fatalf("Len option should be %d, get %d", 5, len(opt))
	}
	_, err := ParseOption(opt)
	if err != nil {
		t.Fatalf("Parsing option error: %s", err.Error())
	}
}
