package models

import (
	"testing"
)

func TestState_FSPath(t *testing.T) {
	tests := []struct {
		name  string
		state *State
		want  string
	}{
		{
			name: "nil",
			state: &State{
				Path: nil,
			},
			want: "/",
		},
		{
			name: "empty",
			state: &State{
				Path: []string{},
			},
			want: "/",
		},
		{
			name: "one",
			state: &State{
				Path: []string{"a"},
			},
			want: "/a",
		},
		{
			name: "two",
			state: &State{
				Path: []string{"a", "b"},
			},
			want: "/a/b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.FSPath(); got != tt.want {
				t.Errorf("State.FSPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
