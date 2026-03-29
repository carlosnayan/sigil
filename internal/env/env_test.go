package env

import (
	"bytes"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	data := []byte(`
# comment
FOO=bar
QUOTED="hello world"
EMPTY=
`)
	m, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if m["FOO"] != "bar" {
		t.Errorf("FOO: got %q", m["FOO"])
	}
	if m["QUOTED"] != "hello world" {
		t.Errorf("QUOTED: got %q", m["QUOTED"])
	}
	if _, ok := m["EMPTY"]; !ok || m["EMPTY"] != "" {
		t.Errorf("EMPTY: got %q ok=%v", m["EMPTY"], ok)
	}
}

func TestParse_empty(t *testing.T) {
	m, err := Parse(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != 0 {
		t.Fatalf("want empty map, got %v", m)
	}
	m2, err := Parse([]byte(""))
	if err != nil {
		t.Fatal(err)
	}
	if len(m2) != 0 {
		t.Fatalf("want empty map, got %v", m2)
	}
}

func TestMerge(t *testing.T) {
	a := map[string]string{"A": "1", "B": "from-a"}
	b := map[string]string{"B": "from-b", "C": "3"}
	got := Merge(a, b)
	want := map[string]string{"A": "1", "B": "from-b", "C": "3"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestSerialize(t *testing.T) {
	m := map[string]string{"Z": "z", "A": "a", "M": "m"}
	b := Serialize(m)
	want := "A=a\nM=m\nZ=z\n"
	if string(b) != want {
		t.Fatalf("got %q want %q", b, want)
	}
}

func TestSerialize_empty(t *testing.T) {
	if got := Serialize(nil); got != nil {
		t.Fatalf("want nil, got %v", got)
	}
	if got := Serialize(map[string]string{}); got != nil {
		t.Fatalf("want nil, got %v", got)
	}
}

func TestSerializeDotEnv(t *testing.T) {
	m := map[string]string{"K": "v"}
	if got := SerializeDotEnv(m); got != "K=v" {
		t.Fatalf("got %q", got)
	}
}

func TestFromSlice(t *testing.T) {
	got := FromSlice([]string{"A=1", "B=c=d", "NOEQ", ""})
	want := map[string]string{"A": "1", "B": "c=d", "NOEQ": ""}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestToSlice(t *testing.T) {
	got := ToSlice(map[string]string{"B": "2", "A": "1"})
	want := []string{"A=1", "B=2"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if ToSlice(nil) != nil {
		t.Fatal("nil map should return nil slice")
	}
	if ToSlice(map[string]string{}) != nil {
		t.Fatal("empty map should return nil slice")
	}
}

func TestRoundTripSerializeParse(t *testing.T) {
	orig := map[string]string{"X": "y", "N": "42"}
	back, err := Parse(Serialize(orig))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(back, orig) {
		t.Fatalf("got %v want %v", back, orig)
	}
}

func TestSerialize_orderingMatchesToSlice(t *testing.T) {
	m := map[string]string{"B": "2", "A": "1"}
	s := Serialize(m)
	slice := ToSlice(m)
	var buf bytes.Buffer
	for _, line := range slice {
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
	if !bytes.Equal(s, buf.Bytes()) {
		t.Fatalf("Serialize vs ToSlice mismatch: %q vs %q", s, buf.Bytes())
	}
}
