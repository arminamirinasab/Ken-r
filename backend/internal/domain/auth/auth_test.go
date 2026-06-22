package auth

import "testing"

func TestNormalizePhone(t *testing.T) {
	cases := map[string]string{
		"09123456789":    "+989123456789",
		"9123456789":     "+989123456789",
		"+989123456789":  "+989123456789",
		"00989123456789": "+989123456789",
		"989123456789":   "+989123456789",
		"0912 345 6789":  "+989123456789",
		"0912-345-6789":  "+989123456789",
	}
	for in, want := range cases {
		got, err := NormalizePhone(in)
		if err != nil {
			t.Errorf("NormalizePhone(%q) error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("NormalizePhone(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizePhoneInvalid(t *testing.T) {
	for _, in := range []string{"", "12345", "08123456789", "091234567890", "+1 555 0100"} {
		if _, err := NormalizePhone(in); err != ErrInvalidPhone {
			t.Errorf("NormalizePhone(%q) err = %v, want ErrInvalidPhone", in, err)
		}
	}
}

func TestNewOTP(t *testing.T) {
	code, err := NewOTP()
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != OTPDigits {
		t.Fatalf("len(code) = %d, want %d", len(code), OTPDigits)
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			t.Fatalf("non-digit in code %q", code)
		}
	}
}
