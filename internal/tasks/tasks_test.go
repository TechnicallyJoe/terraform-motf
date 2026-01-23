package tasks

import (
	"testing"
)

func TestGetShellArgs(t *testing.T) {
	tests := []struct {
		name       string
		shell      string
		script     string
		wantBinary string
		wantArgs   []string
		wantErr    bool
	}{
		{
			name:       "default shell (empty)",
			shell:      "",
			script:     "echo hello",
			wantBinary: "sh",
			wantArgs:   []string{"-c", "echo hello"},
		},
		{
			name:       "sh shell",
			shell:      "sh",
			script:     "echo hello",
			wantBinary: "sh",
			wantArgs:   []string{"-c", "echo hello"},
		},
		{
			name:       "bash shell",
			shell:      "bash",
			script:     "echo hello",
			wantBinary: "bash",
			wantArgs:   []string{"-c", "echo hello"},
		},
		{
			name:       "pwsh shell",
			shell:      "pwsh",
			script:     "Write-Host 'hello'",
			wantBinary: "pwsh",
			wantArgs:   []string{"-Command", "Write-Host 'hello'"},
		},
		{
			name:       "cmd shell",
			shell:      "cmd",
			script:     "echo hello",
			wantBinary: "cmd",
			wantArgs:   []string{"/C", "echo hello"},
		},
		{
			name:    "unknown shell",
			shell:   "zsh",
			script:  "echo hello",
			wantErr: true,
		},
		{
			name:       "multiline script",
			shell:      "sh",
			script:     "echo hello\necho world",
			wantBinary: "sh",
			wantArgs:   []string{"-c", "echo hello\necho world"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binary, args, err := GetShellArgs(tt.shell, tt.script)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if binary != tt.wantBinary {
				t.Errorf("binary = %q, want %q", binary, tt.wantBinary)
			}

			if len(args) != len(tt.wantArgs) {
				t.Fatalf("args length = %d, want %d", len(args), len(tt.wantArgs))
			}

			for i, arg := range args {
				if arg != tt.wantArgs[i] {
					t.Errorf("args[%d] = %q, want %q", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}

func TestSupportedShells(t *testing.T) {
	shells := SupportedShells()

	// Should contain all expected shells
	expected := []string{"bash", "cmd", "pwsh", "sh"}
	if len(shells) != len(expected) {
		t.Errorf("len(shells) = %d, want %d", len(shells), len(expected))
	}

	for i, shell := range shells {
		if shell != expected[i] {
			t.Errorf("shells[%d] = %q, want %q", i, shell, expected[i])
		}
	}
}

func TestNewRunner(t *testing.T) {
	t.Run("with nil tasks", func(t *testing.T) {
		r := NewRunner(nil)
		if r.Tasks == nil {
			t.Error("Tasks should be initialized to empty map, not nil")
		}
	})

	t.Run("with tasks", func(t *testing.T) {
		tasks := map[string]*TaskConfig{
			"test": {Command: "echo test"},
		}
		r := NewRunner(tasks)
		if r.Tasks["test"] == nil {
			t.Error("Tasks should contain the provided tasks")
		}
	})
}

func TestRunner_GetTask(t *testing.T) {
	tasks := map[string]*TaskConfig{
		"hello": {Description: "Say hello", Command: "echo hello"},
	}
	r := NewRunner(tasks)

	t.Run("existing task", func(t *testing.T) {
		task := r.GetTask("hello")
		if task == nil {
			t.Fatal("expected task, got nil")
		}
		if task.Description != "Say hello" {
			t.Errorf("Description = %q, want %q", task.Description, "Say hello")
		}
	})

	t.Run("non-existing task", func(t *testing.T) {
		task := r.GetTask("nonexistent")
		if task != nil {
			t.Error("expected nil for non-existing task")
		}
	})
}

func TestRunner_ListTasks(t *testing.T) {
	tasks := map[string]*TaskConfig{
		"a": {Command: "echo a"},
		"b": {Command: "echo b"},
	}
	r := NewRunner(tasks)

	names := r.ListTasks()
	if len(names) != 2 {
		t.Errorf("len(names) = %d, want 2", len(names))
	}

	// Check both names are present (order not guaranteed)
	hasA, hasB := false, false
	for _, name := range names {
		if name == "a" {
			hasA = true
		}
		if name == "b" {
			hasB = true
		}
	}
	if !hasA || !hasB {
		t.Errorf("expected both 'a' and 'b' in names, got %v", names)
	}
}

func TestRunner_Run_Errors(t *testing.T) {
	t.Run("task not found", func(t *testing.T) {
		r := NewRunner(nil)
		err := r.Run("nonexistent", "/tmp")
		if err == nil {
			t.Error("expected error for non-existing task")
		}
	})

	t.Run("empty command", func(t *testing.T) {
		r := NewRunner(map[string]*TaskConfig{
			"empty": {Description: "No command"},
		})
		err := r.Run("empty", "/tmp")
		if err == nil {
			t.Error("expected error for empty command")
		}
	})

	t.Run("invalid shell", func(t *testing.T) {
		r := NewRunner(map[string]*TaskConfig{
			"bad": {Shell: "zsh", Command: "echo test"},
		})
		err := r.Run("bad", "/tmp")
		if err == nil {
			t.Error("expected error for invalid shell")
		}
	})
}
