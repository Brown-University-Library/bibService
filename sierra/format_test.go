package sierra

import (
	"testing"
)

func TestFormats(t *testing.T) {
	tests := map[string]string{
		"00000nas a2200445 i 4500": "BP",
		"00000cmm a2200457Ma 4500": "CF",
		"00000cjm a2200265Ma 4500": "BSR",
		"00000nem a2200505Ii 4500": "MP",
		"00000ckm a2200445 a 4500": "VM",
	}

	for key, value := range tests {
		if formatCode(key) != value {
			t.Errorf("Incorrect format for: %s", key)
		}
	}
}
