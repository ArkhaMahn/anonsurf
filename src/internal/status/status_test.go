package status

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestReaderSnapshotClassifiesDNSStates(t *testing.T) {
	tests := []struct {
		name           string
		anonsurfActive bool
		addresses      []string
		dnsErr         error
		want           Snapshot
	}{
		{
			name:           "tor dns",
			anonsurfActive: true,
			addresses:      []string{"127.0.0.1"},
			want:           Snapshot{AnonSurfActive: true, TorActive: true, DNSState: DNSTor, DNSAddresses: []string{"127.0.0.1"}},
		},
		{
			name:           "dns leak",
			anonsurfActive: true,
			addresses:      []string{"1.1.1.1"},
			want:           Snapshot{AnonSurfActive: true, TorActive: true, DNSState: DNSLeak, DNSAddresses: []string{"1.1.1.1"}},
		},
		{
			name:      "clearnet dns",
			addresses: []string{"1.1.1.1"},
			want:      Snapshot{TorActive: true, DNSState: DNSClearnet, DNSAddresses: []string{"1.1.1.1"}},
		},
		{
			name:      "localhost only without AnonSurf",
			addresses: []string{"127.0.0.1"},
			want:      Snapshot{TorActive: true, DNSState: DNSLocalhostOnly, DNSAddresses: []string{"127.0.0.1"}},
		},
		{
			name: "empty dns",
			want: Snapshot{TorActive: true, DNSState: DNSEmpty},
		},
		{
			name:   "missing dns",
			dnsErr: os.ErrNotExist,
			want:   Snapshot{TorActive: true, DNSState: DNSMissing},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := Reader{
				AnonSurfActive: func() (bool, error) {
					return test.anonsurfActive, nil
				},
				TorActive: func() (bool, error) {
					return true, nil
				},
				DNSAddresses: func() ([]string, error) {
					return test.addresses, test.dnsErr
				},
			}

			got, err := reader.Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("snapshot = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestReaderSnapshotWrapsReadErrors(t *testing.T) {
	wantErr := errors.New("systemd failed")
	reader := Reader{
		AnonSurfActive: func() (bool, error) {
			return false, wantErr
		},
		TorActive: func() (bool, error) {
			return false, nil
		},
		DNSAddresses: func() ([]string, error) {
			return nil, nil
		},
	}

	_, err := reader.Snapshot()
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want wrapping %v", err, wantErr)
	}
	if !strings.Contains(err.Error(), "read AnonSurf status") {
		t.Fatalf("error = %v, want context", err)
	}
}

func TestReaderSnapshotRejectsMissingReaders(t *testing.T) {
	_, err := (Reader{}).Snapshot()
	if err == nil || !strings.Contains(err.Error(), "missing AnonSurf status reader") {
		t.Fatalf("error = %v, want missing reader error", err)
	}
}
