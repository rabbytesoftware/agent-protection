package rules

import "testing"

func TestRegistry_ParsesEmbeddedYAMLSuccessfully(t *testing.T) {
	reg := Registry()
	if len(reg) == 0 {
		t.Fatal("expected a non-empty registry")
	}
}

func TestRegistry_EveryEntryHasARequiredField(t *testing.T) {
	for id, r := range Registry() {
		if r.ID != id {
			t.Errorf("entry keyed %q has mismatched ID field %q", id, r.ID)
		}
		if r.Text == "" {
			t.Errorf("rule %q has an empty Text", id)
		}
		if r.Rationale == "" {
			t.Errorf("rule %q has an empty Rationale", id)
		}
		if r.FixStrategy == "" {
			t.Errorf("rule %q has an empty FixStrategy", id)
		}
	}
}

func TestRegistry_ContainsAKnownRule(t *testing.T) {
	reg := Registry()
	r, ok := reg["R32"]
	if !ok {
		t.Fatal("expected R32 (no boolean flag parameters) to be present")
	}
	if r.Level != "L0" {
		t.Errorf("R32 Level = %q, want L0", r.Level)
	}
}
