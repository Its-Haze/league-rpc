package process

import (
	"errors"
	"testing"
)

type fakeLister struct {
	names []string
	err   error
}

func (f fakeLister) ProcessNames() ([]string, error) {
	return f.names, f.err
}

func TestChecker_IsRunning(t *testing.T) {
	tests := []struct {
		name    string
		running []string
		check   []string
		want    bool
	}{
		{"exact match", []string{"Discord.exe"}, []string{"Discord.exe"}, true},
		{"case insensitive", []string{"discord.exe"}, []string{"Discord.exe"}, true},
		{"no match", []string{"Explorer.exe"}, []string{"Discord.exe"}, false},
		{"matches any name in the list", []string{"LeagueClientUx.exe"}, []string{"LeagueClient.exe", "LeagueClientUx.exe"}, true},
		{"empty process list", nil, []string{"Discord.exe"}, false},
		{"empty name list", []string{"Discord.exe"}, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCheckerWithLister(fakeLister{names: tt.running})

			got, err := c.IsRunning(tt.check...)
			if err != nil {
				t.Fatalf("IsRunning() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("IsRunning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChecker_IsRunning_ListerError(t *testing.T) {
	wantErr := errors.New("boom")
	c := NewCheckerWithLister(fakeLister{err: wantErr})

	_, err := c.IsRunning("Discord.exe")
	if !errors.Is(err, wantErr) {
		t.Fatalf("IsRunning() error = %v, want %v", err, wantErr)
	}
}
