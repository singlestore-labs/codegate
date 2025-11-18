package codegate

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Code gates are somewhat troublesome to test due to the reliance on a static environment variable
// at the time of the gate creation. All tests run in the same environment, so setting an environment
// variable in one test may affect other tests (depending on order.) DisableGates() testing only
// works because it is implemented to dynamically inspect the environment at call time and does not
// cache results.

func TestNoDisabledGates(t *testing.T) {
	gateTestFoo := New("FOO")
	require.True(t, gateTestFoo.Enabled(), "arbitrary code behavior should be enabled by default")
	require.NotContains(t, DisabledGates(true), "FOO")
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
	_ = os.Setenv("DISABLE_Bar", "disabled")

	// refresh disabled gates to pick up the changes to the environment
	// variables
	DisabledGates(true)

	gateTestBar := New("Bar")
	require.False(t, gateTestBar.Enabled(), "Bar should be disabled")
	require.True(t, New("Bar2").Enabled(), "Other gates should be enabled")
	require.Contains(t, DisabledGates(false), "Bar")
}

func TestDisableMultipleGates(t *testing.T) {
	// create some random environment variables
	_ = os.Setenv("NOISE", "LOUD")
	_ = os.Setenv("MORE_NOISE", "LOUDER")
	_ = os.Setenv("DISABLED_CODE", "")

	// disable two gates
	_ = os.Setenv("DISABLE_Baz1", "disabled")
	_ = os.Setenv("DISABLE_Baz3", "disabled")

	// refresh disabled gates to pick up the changes to the environment
	// variables
	DisabledGates(true)

	// define four gates
	gateTestBaz1 := New("Baz1")
	gateTestBaz2 := New("Baz2")
	gateTestBaz3 := New("Baz3")
	gateTestBaz4 := New("Baz4")

	require.False(t, gateTestBaz1.Enabled(), "Baz1 should be disabled")
	require.True(t, gateTestBaz2.Enabled(), "Baz2 should be enabled")
	require.False(t, gateTestBaz3.Enabled(), "Baz3 should be disabled")
	require.True(t, gateTestBaz4.Enabled(), "Baz4 should be enabled")

	require.Contains(t, DisabledGates(false), "Baz1")
	require.NotContains(t, DisabledGates(false), "Baz2")
	require.Contains(t, DisabledGates(false), "Baz3")
	require.NotContains(t, DisabledGates(false), "Baz4")
}

func TestRefreshDisabledGates(t *testing.T) {
	// ensure no disabled gates at start
	_ = os.Unsetenv("DISABLE_Foo")
	require.NotContains(t, DisabledGates(false), "Foo")

	// disable Foo
	_ = os.Setenv("DISABLE_Foo", "disabled")
	require.NotContains(t, DisabledGates(false), "Foo")
	// refresh disabled gates
	require.Contains(t, DisabledGates(true), "Foo", "DisabledGates(true) should refresh the disabled gates")
}
