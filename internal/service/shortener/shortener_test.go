package shortener

import (
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func TestGenerateID(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "#1 - generate random ID",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateID()

			assert.NoError(t, err, "GenerateID() returned unexpected error")

			assert.NotEmpty(t, got, "GenerateID() returned empty string")

			assert.Len(t, got, 8, "GenerateID() should return 8-character string")

			for _, c := range got {
				assert.Truef(t, unicode.IsLetter(c) || unicode.IsDigit(c),
					"GenerateID() returned invalid character: %q", c)
			}
		})
	}
}
