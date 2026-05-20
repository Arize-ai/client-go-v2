package optfields_test

import (
	"testing"
	"time"

	"github.com/Arize-ai/client-go-v2/arize/internal/optfields"
)

func TestPtrIfSet_String(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantNil bool
	}{
		{name: "empty returns nil", in: "", wantNil: true},
		{name: "non-empty returns pointer", in: "demo", wantNil: false},
		{name: "whitespace is non-empty", in: " ", wantNil: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := optfields.PtrIfSet(tt.in)
			if tt.wantNil {
				if got != nil {
					t.Fatalf("PtrIfSet(%q) = %v, want nil", tt.in, *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("PtrIfSet(%q) = nil, want non-nil", tt.in)
			}
			if *got != tt.in {
				t.Fatalf("*PtrIfSet(%q) = %q, want %q", tt.in, *got, tt.in)
			}
		})
	}
}

func TestPtrIfSet_Int(t *testing.T) {
	tests := []struct {
		name    string
		in      int
		wantNil bool
	}{
		{name: "zero returns nil", in: 0, wantNil: true},
		{name: "positive returns pointer", in: 50, wantNil: false},
		{name: "negative returns pointer", in: -1, wantNil: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := optfields.PtrIfSet(tt.in)
			if tt.wantNil {
				if got != nil {
					t.Fatalf("PtrIfSet(%d) = %v, want nil", tt.in, *got)
				}
				return
			}
			if got == nil || *got != tt.in {
				t.Fatalf("PtrIfSet(%d) returned wrong value", tt.in)
			}
		})
	}
}

func TestPtrIfSet_Time(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		in      time.Time
		wantNil bool
	}{
		{name: "zero time returns nil", in: time.Time{}, wantNil: true},
		{name: "concrete time returns pointer", in: now, wantNil: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := optfields.PtrIfSet(tt.in)
			if tt.wantNil {
				if got != nil {
					t.Fatalf("PtrIfSet(%v) = %v, want nil", tt.in, *got)
				}
				return
			}
			if got == nil || !got.Equal(tt.in) {
				t.Fatalf("PtrIfSet(%v) returned wrong value", tt.in)
			}
		})
	}
}

func TestPtrIfSet_Enum(t *testing.T) {
	type status string
	const (
		statusActive status = "ACTIVE"
	)
	tests := []struct {
		name    string
		in      status
		wantNil bool
	}{
		{name: "empty enum returns nil", in: status(""), wantNil: true},
		{name: "valid enum returns pointer", in: statusActive, wantNil: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := optfields.PtrIfSet(tt.in)
			if tt.wantNil {
				if got != nil {
					t.Fatalf("PtrIfSet(%q) = %v, want nil", tt.in, *got)
				}
				return
			}
			if got == nil || *got != tt.in {
				t.Fatalf("PtrIfSet(%q) returned wrong value", tt.in)
			}
		})
	}
}

func TestPtrWithDefault(t *testing.T) {
	tests := []struct {
		name     string
		in       int
		fallback int
		want     int
	}{
		{name: "zero substitutes fallback", in: 0, fallback: 50, want: 50},
		{name: "non-zero keeps value", in: 10, fallback: 50, want: 10},
		{name: "negative is non-zero", in: -1, fallback: 50, want: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := optfields.PtrWithDefault(tt.in, tt.fallback)
			if got == nil {
				t.Fatalf("PtrWithDefault(%d, %d) = nil, want non-nil", tt.in, tt.fallback)
			}
			if *got != tt.want {
				t.Fatalf("PtrWithDefault(%d, %d) = %d, want %d", tt.in, tt.fallback, *got, tt.want)
			}
		})
	}
}

func TestPtrWithDefault_String(t *testing.T) {
	got := optfields.PtrWithDefault("", "fallback")
	if got == nil || *got != "fallback" {
		t.Fatalf("PtrWithDefault(\"\", \"fallback\") returned wrong value")
	}
	got = optfields.PtrWithDefault("hello", "fallback")
	if got == nil || *got != "hello" {
		t.Fatalf("PtrWithDefault(\"hello\", \"fallback\") returned wrong value")
	}
}
