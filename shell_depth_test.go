package autocd

import (
	"os"
	"strings"
	"testing"
)

// Test shell depth detection on Unix systems
func TestCheckShellDepth_Unix(t *testing.T) {

	tests := []struct {
		name             string
		shlvl            string
		threshold        int
		disableWarnings  bool
		expectWarning    bool
		expectedContains []string
	}{
		{
			name:            "below_threshold",
			shlvl:           "5",
			threshold:       15,
			disableWarnings: false,
			expectWarning:   false,
		},
		{
			name:             "at_threshold",
			shlvl:            "15",
			threshold:        15,
			disableWarnings:  false,
			expectWarning:    true,
			expectedContains: []string{"💡 Tip: You have 15 nested shells", "For better performance"},
		},
		{
			name:             "above_threshold",
			shlvl:            "25",
			threshold:        15,
			disableWarnings:  false,
			expectWarning:    true,
			expectedContains: []string{"💡 Tip: You have 25 nested shells", "For better performance"},
		},
		{
			name:             "custom_threshold",
			shlvl:            "8",
			threshold:        8,
			disableWarnings:  false,
			expectWarning:    true,
			expectedContains: []string{"💡 Tip: You have 8 nested shells"},
		},
		{
			name:            "warnings_disabled",
			shlvl:           "30",
			threshold:       15,
			disableWarnings: true,
			expectWarning:   false,
		},
		{
			name:          "no_shlvl",
			shlvl:         "",
			threshold:     15,
			expectWarning: false,
		},
		{
			name:          "invalid_shlvl",
			shlvl:         "invalid",
			threshold:     15,
			expectWarning: false,
		},
		{
			name:          "negative_shlvl",
			shlvl:         "-5",
			threshold:     15,
			expectWarning: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original SHLVL
			originalShlvl := os.Getenv("SHLVL")
			defer func() {
				if originalShlvl == "" {
					os.Unsetenv("SHLVL")
				} else {
					os.Setenv("SHLVL", originalShlvl)
				}
			}()

			// Set test SHLVL
			if tt.shlvl == "" {
				os.Unsetenv("SHLVL")
			} else {
				os.Setenv("SHLVL", tt.shlvl)
			}

			// Capture stderr output
			originalStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			// Create test options
			opts := &Options{
				DepthWarningThreshold: tt.threshold,
				DisableDepthWarnings:  tt.disableWarnings,
			}

			// Run the function
			checkShellDepth(opts)

			// Restore stderr and read output
			w.Close()
			os.Stderr = originalStderr
			output := make([]byte, 1024)
			n, _ := r.Read(output)
			stderrOutput := string(output[:n])

			// Check expectations
			if tt.expectWarning {
				if stderrOutput == "" {
					t.Error("Expected warning output but got none")
				}
				for _, expected := range tt.expectedContains {
					if !strings.Contains(stderrOutput, expected) {
						t.Errorf("Expected output to contain '%s', got: %s", expected, stderrOutput)
					}
				}
			} else {
				if stderrOutput != "" {
					t.Errorf("Expected no warning output but got: %s", stderrOutput)
				}
			}
		})
	}
}

// Test Options struct defaults for shell depth fields
func TestOptions_ShellDepthDefaults(t *testing.T) {
	tests := []struct {
		name                    string
		opts                    *Options
		expectedThreshold       int
		expectedDisableWarnings bool
	}{
		{
			name:                    "nil_options",
			opts:                    nil,
			expectedThreshold:       15,
			expectedDisableWarnings: false,
		},
		{
			name:                    "zero_threshold",
			opts:                    &Options{DepthWarningThreshold: 0},
			expectedThreshold:       15,
			expectedDisableWarnings: false,
		},
		{
			name:                    "custom_threshold",
			opts:                    &Options{DepthWarningThreshold: 10},
			expectedThreshold:       10,
			expectedDisableWarnings: false,
		},
		{
			name:                    "warnings_disabled",
			opts:                    &Options{DisableDepthWarnings: true},
			expectedThreshold:       15,
			expectedDisableWarnings: true,
		},
		{
			name: "full_custom",
			opts: &Options{
				DepthWarningThreshold: 5,
				DisableDepthWarnings:  true,
			},
			expectedThreshold:       5,
			expectedDisableWarnings: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the default setting logic from ExitWithDirectoryAdvanced
			opts := tt.opts
			if opts == nil {
				opts = &Options{
					SecurityLevel:         SecurityNormal,
					DebugMode:             os.Getenv("AUTOCD_DEBUG") != "",
					DepthWarningThreshold: 15,
					DisableDepthWarnings:  false,
				}
			}

			// Set defaults for new fields if not specified
			if opts.DepthWarningThreshold == 0 {
				opts.DepthWarningThreshold = 15
			}

			// Check expectations
			if opts.DepthWarningThreshold != tt.expectedThreshold {
				t.Errorf("Expected threshold %d, got %d", tt.expectedThreshold, opts.DepthWarningThreshold)
			}
			if opts.DisableDepthWarnings != tt.expectedDisableWarnings {
				t.Errorf("Expected DisableDepthWarnings %v, got %v", tt.expectedDisableWarnings, opts.DisableDepthWarnings)
			}
		})
	}
}

// Test integration with main AutoCD functions (mock test since we can't actually exec)
func TestShellDepthIntegration(t *testing.T) {

	// Save original SHLVL
	originalShlvl := os.Getenv("SHLVL")
	defer func() {
		if originalShlvl == "" {
			os.Unsetenv("SHLVL")
		} else {
			os.Setenv("SHLVL", originalShlvl)
		}
	}()

	// Set high SHLVL to trigger warning
	os.Setenv("SHLVL", "20")

	// Capture stderr output
	originalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Try to call ExitWithDirectory (it will fail at path validation, but shell depth check should run first)
	err := ExitWithDirectory("/nonexistent/path/for/testing")

	// Restore stderr and read output
	w.Close()
	os.Stderr = originalStderr
	output := make([]byte, 1024)
	n, _ := r.Read(output)
	stderrOutput := string(output[:n])

	// Should get both shell depth warning AND path validation error
	if err == nil {
		t.Error("Expected error from invalid path")
	}

	if !strings.Contains(stderrOutput, "💡 Tip: You have 20 nested shells") {
		t.Error("Expected shell depth warning in stderr output")
	}
}

// Benchmark shell depth checking performance
func BenchmarkCheckShellDepth(b *testing.B) {
	opts := &Options{
		DepthWarningThreshold: 15,
		DisableDepthWarnings:  false,
	}

	// Set SHLVL for Unix systems
	os.Setenv("SHLVL", "10")
	defer os.Unsetenv("SHLVL")

	// Redirect stderr to avoid cluttering benchmark output
	originalStderr := os.Stderr
	devNull, _ := os.Open(os.DevNull)
	os.Stderr = devNull
	defer func() {
		os.Stderr = originalStderr
		devNull.Close()
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checkShellDepth(opts)
	}
}
