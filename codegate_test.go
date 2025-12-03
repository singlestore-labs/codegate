package codegate

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Code gates are troublesome to test due to the reliance on static environment
// variables at the time of the gate creation. All tests run in the same environment,
// so setting an environment in one test may affect other tests. All tests in this
// file should use unique gate names to avoid conflicts.

func TestNoDisabledGates(t *testing.T) {
	gateTestFoo := New("FOO")
	require.True(t, gateTestFoo.Enabled(), "arbitrary code behavior should be enabled by default")
	require.NotContains(t, DisabledGates(), "FOO")
}

func TestGateNames(t *testing.T) {
	// valid names
	_ = New("A")
	_ = New("Z123")
	_ = New("A_B______C_D____")
	_ = New("ABCDEFGHIJKLMNOPQRSTUVWXYZ_0123456789")
	_ = New("RBACDeleteOrphanedGrants")
	_ = New("RBAC_DELETE_ORPHANED_GRANTS")
	_ = New(strings.Repeat("Z", 100))

	// invalid names
	require.Panics(t, func() { New("") })
	require.Panics(t, func() { New("1") })
	require.Panics(t, func() { New("%") })
	require.Panics(t, func() { New("A$") })
	require.Panics(t, func() { New("A-A") })
	require.Panics(t, func() {
		New(strings.Repeat("A", 101))
	})

	// duplicate name
	require.Panics(t, func() { New("A") })
}

func TestDisableOneGate(t *testing.T) {
	t.Setenv("DISABLE_CODE_Bar", "disabled")

	// refresh disabled gates to pick up the changes to the environment
	// variables
	resetDisabledGates()

	gateTestBar := New("Bar")
	require.False(t, gateTestBar.Enabled(), "Bar should be disabled")
	require.True(t, New("Bar2").Enabled(), "Other gates should be enabled")
	require.Contains(t, DisabledGates(), "Bar")
}

func TestDisableOldGate(t *testing.T) {
	t.Setenv("DISABLE_S2CODE_Deprecated", "disabled")

	// refresh disabled gates to pick up the changes to the environment
	// variables
	resetDisabledGates()

	gateTestDeprecated := New("Deprecated")
	require.False(t, gateTestDeprecated.Enabled(), "Deprecated should be disabled")
	require.True(t, New("Deprecated2").Enabled(), "Other gates should be enabled")
	require.Contains(t, DisabledGates(), "Deprecated")
}

func TestDisableMultipleGates(t *testing.T) {
	// create some random environment variables
	t.Setenv("NOISE", "LOUD")
	t.Setenv("MORE_NOISE", "LOUDER")
	t.Setenv("DISABLED_CODE", "")
	t.Setenv("DISABLED_CODE_TOO", "")

	// disable two gates
	t.Setenv("DISABLE_CODE_Baz1", "disabled")
	t.Setenv("DISABLE_CODE_Baz3", "disabled")

	// refresh disabled gates to pick up the changes to the environment
	// variables
	resetDisabledGates()

	// define four gates
	gateTestBaz1 := New("Baz1")
	gateTestBaz2 := New("Baz2")
	gateTestBaz3 := New("Baz3")
	gateTestBaz4 := New("Baz4")

	require.False(t, gateTestBaz1.Enabled(), "Baz1 should be disabled")
	require.True(t, gateTestBaz2.Enabled(), "Baz2 should be enabled")
	require.False(t, gateTestBaz3.Enabled(), "Baz3 should be disabled")
	require.True(t, gateTestBaz4.Enabled(), "Baz4 should be enabled")

	require.Contains(t, DisabledGates(), "Baz1")
	require.NotContains(t, DisabledGates(), "Baz2")
	require.Contains(t, DisabledGates(), "Baz3")
	require.NotContains(t, DisabledGates(), "Baz4")
}

func TestResetDisabledGates(t *testing.T) {
	// ensure no disabled gates at start
	require.NotContains(t, DisabledGates(), "Foo")

	// disable Foo
	t.Setenv("DISABLE_CODE_Foo", "disabled")
	require.NotContains(t, DisabledGates(), "Foo")
	// refresh disabled gates
	resetDisabledGates()
	require.Contains(t, DisabledGates(), "Foo", "DisabledGates(true) should refresh the disabled gates")
}
