package regsrc

import (
	"testing"
)

func TestNewTerraformProviderNamespace(t *testing.T) {
	tests := map[string]struct {
		name              string
		provider          string
		source            string
		expectedNamespace string
		expectedName      string
	}{
		"default": {
			provider:          "null",
			source:            "null",
			expectedNamespace: "-",
		},
		"explicit": {
			name:              "explicit",
			provider:          "null",
			source:            "terraform-providers/null",
			expectedNamespace: "terraform-providers",
		},

		"community": {
			name:              "community",
			provider:          "null",
			source:            "community-providers/null",
			expectedNamespace: "community-providers",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actual := NewTerraformProvider(tt.provider, "", "", tt.source)
			if actual == nil {
				t.Fatal("NewTerraformProvider() unexpectedly returned nil provider")
			}
			if v := actual.RawNamespace; v != tt.expectedNamespace {
				t.Fatalf("RawNamespace = %v, wanted %v", v, tt.expectedNamespace)
			}
		})
	}
}
