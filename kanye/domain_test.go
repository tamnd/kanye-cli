package kanye

import "testing"

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "kanye" {
		t.Errorf("Scheme = %q, want kanye", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "kanye" {
		t.Errorf("Binary = %q, want kanye", info.Identity.Binary)
	}
}

func TestClassify(t *testing.T) {
	_, _, err := Domain{}.Classify("")
	if err == nil {
		t.Error("Classify empty string should return error")
	}

	typ, id, err := Domain{}.Classify("some-quote")
	if err != nil {
		t.Errorf("Classify: unexpected error: %v", err)
	}
	if typ != "quote" {
		t.Errorf("Classify type = %q, want quote", typ)
	}
	if id != "some-quote" {
		t.Errorf("Classify id = %q, want some-quote", id)
	}
}

func TestLocate(t *testing.T) {
	got, err := Domain{}.Locate("quote", "anything")
	if err != nil {
		t.Fatalf("Locate: unexpected error: %v", err)
	}
	if got == "" {
		t.Error("Locate returned empty URL")
	}

	_, err = Domain{}.Locate("unknown", "x")
	if err == nil {
		t.Error("Locate with unknown type should return error")
	}
}
