package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestSubFilter(t *testing.T) {
	incoming := []byte{127, 128, 139, 134, 139, 133, 136, 129}
	expected := []byte{127, 1, 11, 251, 5, 250, 3, 249}
	got := subFilter(incoming)

	if bytes.Compare(expected, got) != 0 {
		t.Fatalf("expected %v, but got %v", expected, got)
	}
}

func TestUnSubFilter(t *testing.T) {
	original := []byte{127, 128, 139, 134, 139, 133, 136, 129}
	// expected := []byte{127, 1, 11, 251, 5, 250, 3, 249}
	filtered := subFilter(original)
	fmt.Println("filtered", filtered)

	got := unSubFilter(filtered)

	if bytes.Compare(original, got) != 0 {
		t.Fatalf("expected %v, but got %v", original, got)
	}
}

func TestUpFilter(t *testing.T) {
	prev := []byte{127, 128, 139, 134, 139, 133, 136, 129}
	incoming := []byte{131, 132, 129, 131, 138, 139, 137, 139}
	expected := []byte{4, 4, 246, 253, 255, 6, 1, 10}
	got := upFilter(prev, incoming)

	if bytes.Compare(expected, got) != 0 {
		t.Fatalf("expected %v, but got %v", expected, got)
	}
}

func TestUnUpFilter(t *testing.T) {
	prev := []byte{127, 128, 139, 134, 139, 133, 136, 129}
	original := []byte{131, 132, 129, 131, 138, 139, 137, 139}

	filtered := upFilter(prev, original)
	// expected := []byte{4, 4, 246, 253, 255, 6, 1, 10}
	got := unUpFilter(prev, filtered)

	if bytes.Compare(original, got) != 0 {
		t.Fatalf("expected %v, but got %v", original, got)
	}
}

func TestPaethFilter(t *testing.T) {
	t.Skip("skipping peath filter atm")
	prev := []byte{129, 131, 134, 134, 136, 128, 134, 127}
	incoming := []byte{133, 136, 136, 134, 133, 136, 128, 127}
	expected := []byte{4, 3, 0, 254, 253, 8, 248, 0}
	got := paethFilter(prev, incoming)

	if bytes.Compare(expected, got) != 0 {
		t.Fatalf("expected %v, but got %v", expected, got)
	}
}
