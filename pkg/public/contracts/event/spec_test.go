package event

import "testing"

func TestLookupFindsDevicePropertyReported(t *testing.T) {
	spec, ok := Lookup(TypeDevicePropertyReported, VersionDevicePropertyReported)
	if !ok {
		t.Fatal("Lookup() ok = false")
	}

	if spec != DevicePropertyReported {
		t.Fatalf("Lookup() = %#v", spec)
	}
}

func TestLookupReturnsFalseForUnknownEvent(t *testing.T) {
	if _, ok := Lookup("device.property.reported", 2); ok {
		t.Fatal("Lookup() ok = true, want false")
	}
}

func TestSpecKey(t *testing.T) {
	got := DevicePropertyReported.Key()
	want := PayloadSchemaKey(TypeDevicePropertyReported, VersionDevicePropertyReported)

	if got != want {
		t.Fatalf("Key() = %q, want %q", got, want)
	}
}

func TestSpecsReturnsCopy(t *testing.T) {
	got := Specs()
	got[0] = Spec{}

	again := Specs()
	if again[0] != DevicePropertyReported {
		t.Fatalf("Specs() returned mutable backing slice: %#v", again[0])
	}
}

func TestCatalogSpecsAreValidAndUnique(t *testing.T) {
	seen := make(map[string]bool)

	for _, spec := range Specs() {
		if spec.Type == "" {
			t.Fatal("catalog spec has empty Type")
		}
		if spec.Version < 1 {
			t.Fatalf("catalog spec %q has invalid Version %d", spec.Type, spec.Version)
		}
		if spec.SchemaFile == "" {
			t.Fatalf("catalog spec %s v%d has empty SchemaFile", spec.Type, spec.Version)
		}
		if spec.Topic == "" {
			t.Fatalf("catalog spec %s v%d has empty Topic", spec.Type, spec.Version)
		}

		key := spec.Key()
		if seen[key] {
			t.Fatalf("duplicate catalog spec key %q", key)
		}
		seen[key] = true
	}
}
