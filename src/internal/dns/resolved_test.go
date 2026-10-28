package dns

import "testing"

func TestParseActiveState(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		failed  bool
		want    bool
		wantErr bool
	}{
		{name: "active", output: "active\n", want: true},
		{name: "activating", output: "activating\n", want: true},
		{name: "reloading", output: "reloading\n", want: true},
		{name: "inactive", output: "inactive\n", failed: true},
		{name: "failed", output: "failed\n", failed: true},
		{name: "unknown", output: "unknown\n", failed: true},
		{name: "deactivating", output: "deactivating\n", failed: true},
		{name: "empty output", failed: true},
		{name: "unexpected state", output: "bogus\n", failed: true, wantErr: true},
		{name: "unexpected success state", output: "bogus\n", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseActiveState(test.output, test.failed)
			if test.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got state=%v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != test.want {
				t.Fatalf("expected active=%v, got %v", test.want, got)
			}
		})
	}
}
