package startup

import (
	"errors"
	"testing"
)

const testExe = `C:\Apps\league-rpc-gui.exe`

// fakeRunKey stands in for the HKCU Run key. No real registry is touched.
type fakeRunKey struct {
	values  map[string]string
	sets    int
	deletes int
	getErr  error
}

func newFakeRunKey() *fakeRunKey { return &fakeRunKey{values: map[string]string{}} }

func (f *fakeRunKey) Value(name string) (string, bool, error) {
	if f.getErr != nil {
		return "", false, f.getErr
	}
	v, ok := f.values[name]
	return v, ok, nil
}

func (f *fakeRunKey) SetValue(name, value string) error {
	f.sets++
	f.values[name] = value
	return nil
}

func (f *fakeRunKey) DeleteValue(name string) error {
	f.deletes++
	delete(f.values, name)
	return nil
}

func testReconciler(key RunKey) *Reconciler {
	r := New(key)
	r.exePath = func() (string, error) { return testExe, nil }
	return r
}

func TestReconcile_EnableWritesValue(t *testing.T) {
	key := newFakeRunKey()
	r := testReconciler(key)

	if err := r.Reconcile(true); err != nil {
		t.Fatalf("Reconcile(true): %v", err)
	}

	want := `"` + testExe + `" ` + HiddenArg
	if got := key.values[ValueName]; got != want {
		t.Fatalf("Run value = %q, want %q", got, want)
	}
	if key.sets != 1 {
		t.Fatalf("expected exactly one write, got %d", key.sets)
	}
	if on, _ := r.Enabled(); !on {
		t.Fatal("Enabled() should report true after enable")
	}
}

func TestReconcile_DisableRemovesValue(t *testing.T) {
	key := newFakeRunKey()
	key.values[ValueName] = Command(testExe)
	r := testReconciler(key)

	if err := r.Reconcile(false); err != nil {
		t.Fatalf("Reconcile(false): %v", err)
	}

	if _, ok := key.values[ValueName]; ok {
		t.Fatal("Run value should be gone after disable")
	}
	if key.deletes != 1 {
		t.Fatalf("expected exactly one delete, got %d", key.deletes)
	}
}

func TestReconcile_Idempotent(t *testing.T) {
	key := newFakeRunKey()
	r := testReconciler(key)

	for i := range 3 {
		if err := r.Reconcile(true); err != nil {
			t.Fatalf("Reconcile(true) #%d: %v", i, err)
		}
	}
	if key.sets != 1 {
		t.Fatalf("enable should write once across repeats, wrote %d times", key.sets)
	}

	for i := range 3 {
		if err := r.Reconcile(false); err != nil {
			t.Fatalf("Reconcile(false) #%d: %v", i, err)
		}
	}
	if key.deletes != 1 {
		t.Fatalf("disable should delete once across repeats, deleted %d times", key.deletes)
	}
}

func TestReconcile_DisableWhenAbsentIsNoop(t *testing.T) {
	key := newFakeRunKey()
	r := testReconciler(key)

	if err := r.Reconcile(false); err != nil {
		t.Fatalf("Reconcile(false): %v", err)
	}
	if key.deletes != 0 {
		t.Fatalf("nothing to remove, but delete was called %d times", key.deletes)
	}
}

func TestReconcile_RewritesStaleCommand(t *testing.T) {
	key := newFakeRunKey()
	key.values[ValueName] = `"C:\Old\path.exe"`
	r := testReconciler(key)

	if err := r.Reconcile(true); err != nil {
		t.Fatalf("Reconcile(true): %v", err)
	}
	if got, want := key.values[ValueName], Command(testExe); got != want {
		t.Fatalf("stale command not rewritten: got %q, want %q", got, want)
	}
	if key.sets != 1 {
		t.Fatalf("expected one rewrite, got %d", key.sets)
	}
}

func TestReconcile_PropagatesReadError(t *testing.T) {
	key := newFakeRunKey()
	key.getErr = errors.New("registry unavailable")
	r := testReconciler(key)

	if err := r.Reconcile(true); err == nil {
		t.Fatal("expected the read error to surface")
	}
	if key.sets != 0 || key.deletes != 0 {
		t.Fatal("a failed read must not lead to a write or delete")
	}
}

func TestStartedHidden(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"marker present", []string{HiddenArg}, true},
		{"marker among others", []string{"--foo", HiddenArg, "bar"}, true},
		{"no args", nil, false},
		{"other args only", []string{"--foo", "bar"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StartedHidden(tt.args); got != tt.want {
				t.Fatalf("StartedHidden(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestCommand(t *testing.T) {
	got := Command(`C:\Program Files\league-rpc\app.exe`)
	want := `"C:\Program Files\league-rpc\app.exe" --hidden`
	if got != want {
		t.Fatalf("Command() = %q, want %q", got, want)
	}
}
