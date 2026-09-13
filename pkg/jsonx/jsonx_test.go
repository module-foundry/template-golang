package jsonx

import (
	"testing"
)

type payload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	in := payload{Name: "ada", Age: 36}
	data, err := Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out payload
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out != in {
		t.Fatalf("round trip mismatch: %+v", out)
	}
}

func TestUnmarshalStrictRejectsUnknownMembers(t *testing.T) {
	var out payload
	if err := UnmarshalStrict([]byte(`{"name":"ada","age":36,"extra":true}`), &out); err == nil {
		t.Fatal("strict unmarshal must reject unknown members")
	}
}

func TestUnmarshalStrictAllowsKnownMembers(t *testing.T) {
	var out payload
	if err := UnmarshalStrict([]byte(`{"name":"ada","age":36}`), &out); err != nil {
		t.Fatalf("strict unmarshal: %v", err)
	}
}

func TestFiberAdapters(t *testing.T) {
	data, err := FiberEncoder(payload{Name: "ada"})
	if err != nil {
		t.Fatalf("encoder: %v", err)
	}
	var out payload
	if err := FiberDecoder(data, &out); err != nil {
		t.Fatalf("decoder: %v", err)
	}
	if out.Name != "ada" {
		t.Fatalf("name = %q", out.Name)
	}
}

func BenchmarkMarshal(b *testing.B) {
	in := payload{Name: "ada", Age: 36}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := Marshal(in); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	data := []byte(`{"name":"ada","age":36}`)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var out payload
		if err := Unmarshal(data, &out); err != nil {
			b.Fatal(err)
		}
	}
}
