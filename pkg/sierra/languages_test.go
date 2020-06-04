package sierra

import (
	"testing"
)

func TestLanguageName(t *testing.T) {
	name := languageName("")
	if name != "" {
		t.Errorf("Invalid language name (empty case): %#v", name)
	}

	name = languageName("eng")
	if name != "English" {
		t.Errorf("Invalid language name (single case): %#v", name)
	}

	names := languageNames("engchi")
	if names[0] != "English" || names[1] != "Chinese" {
		t.Errorf("Invalid language names (multi-case): %#v", names)
	}

	names = languageNames("engch")
	if len(names) != 1 || names[0] != "English" {
		t.Errorf("Invalid languages returned: %#v", names)
	}

	names = languageNames("")
	if len(names) != 0 {
		t.Errorf("Invalid languages returned: %#v", names)
	}
}
