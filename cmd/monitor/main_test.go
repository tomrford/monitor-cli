package main

import "testing"

func TestNormalizeCadence(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "hourly", input: "hourly", want: "hourly"},
		{name: "daily", input: "daily", want: "daily"},
		{name: "weekly", input: "weekly", want: "weekly"},
		{name: "trim", input: "  daily ", want: "daily"},
		{name: "lower", input: "WEEKLY", want: "weekly"},
		{name: "invalid", input: "every_two_weeks", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizeCadence(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
