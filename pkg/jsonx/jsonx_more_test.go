package jsonx

import (
	"strings"
	"testing"
)

func TestUnmarshalInvalidJSON(t *testing.T) {
	var out payload
	if err := Unmarshal([]byte("{"), &out); err == nil {
		t.Fatal("expected error for malformed json")
	}
}

func TestUnmarshalStrictInvalidJSON(t *testing.T) {
	var out payload
	if err := UnmarshalStrict([]byte("[]"), &out); err == nil {
		t.Fatal("expected error for wrong json type")
	}
}

func TestMarshalUnsupportedValue(t *testing.T) {
	if _, err := Marshal(make(chan int)); err == nil {
		t.Fatal("expected error for unsupported value")
	}
}

func TestFiberDecoderInvalid(t *testing.T) {
	var out payload
	if err := FiberDecoder([]byte("{"), &out); err == nil {
		t.Fatal("expected decoder error")
	}
}

func TestMarshalDeterministic(t *testing.T) {
	first, err := MarshalDeterministic(map[string]int{"b": 2, "a": 1})
	if err != nil {
		t.Fatal(err)
	}
	second, err := MarshalDeterministic(map[string]int{"b": 2, "a": 1})
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) || !strings.HasPrefix(string(first), `{"a"`) {
		t.Fatalf("not deterministic: %s / %s", first, second)
	}
}
