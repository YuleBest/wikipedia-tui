package app

import "testing"

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"LC_ALL wins", map[string]string{"LC_ALL": "zh_CN.UTF-8", "LANG": "en_US.UTF-8"}, "zh"},
		{"falls back to LC_MESSAGES", map[string]string{"LC_ALL": "C", "LC_MESSAGES": "ja_JP.UTF-8"}, "ja"},
		{"falls back to LANG", map[string]string{"LANG": "de_DE@euro"}, "de"},
		{"defaults to English", map[string]string{"LANG": "C"}, "en"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectLanguage(func(key string) string { return tt.env[key] })
			if got != tt.want {
				t.Fatalf("DetectLanguage() = %q, want %q", got, tt.want)
			}
		})
	}
}
