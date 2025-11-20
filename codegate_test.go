package codegate_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/singlestore-labs/codegate"
)

// Code gates are somewhat troublesome to test due to the reliance on static environment variables
// at the time of the gate creation. All tests run in the same environment, so setting an environment
// variable in one test may affect other tests (depending on order.) Gate names must be unique
// across all tests.

func TestNoDisabledGates(t *testing.T) {
	gateTestFoo := codegate.New("FOO")
	require.True(t, gateTestFoo.Enabled(), "arbitrary code behavior should be enabled by default")
	require.NotContains(t, codegate.DisabledGates(), "FOO")
}

func TestGateNames(t *testing.T) {
	// valid names
	_ = codegate.New("A")
	_ = codegate.New("Z123")
	_ = codegate.New("A_B______C_D____")
	_ = codegate.New("ABCDEFGHIJKLMNOPQRSTUVWXYZ_0123456789")
	_ = codegate.New("RBACDeleteOrphanedGrants")
	_ = codegate.New("RBAC_DELETE_ORPHANED_GRANTS")
	_ = codegate.New(strings.Repeat("Z", 100))

	// invalid names
	require.Panics(t, func() { codegate.New("") })
	require.Panics(t, func() { codegate.New("1") })
	require.Panics(t, func() { codegate.New("%") })
	require.Panics(t, func() { codegate.New("A$") })
	require.Panics(t, func() { codegate.New("A-A") })
	require.Panics(t, func() {
		codegate.New(strings.Repeat("A", 101))
	})

	// duplicate name
	require.Panics(t, func() { codegate.New("A") })
}

func TestDisableOneGate(t *testing.T) {
	_ = os.Setenv("DISABLE_CODE_Bar", "disabled")

	gateTestBar := codegate.New("Bar")
	require.False(t, gateTestBar.Enabled(), "Bar should be disabled")
	require.True(t, codegate.New("Bar2").Enabled(), "Other gates should be enabled")
	require.Contains(t, codegate.DisabledGates(), "Bar")
	require.NotContains(t, codegate.DisabledGates(), "Bar2")
}

func TestDisableMultipleGates(t *testing.T) {
	// create some random environment variables
	_ = os.Setenv("NOISE", "LOUD")
	_ = os.Setenv("MORE_NOISE", "LOUDER")
	_ = os.Setenv("DISABLED_CODE", "")
	_ = os.Setenv("DISABLED_CODE_TOO", "")

	// disable two gates
	_ = os.Setenv("DISABLE_CODE_Baz1", "disabled")
	_ = os.Setenv("DISABLE_CODE_Baz3", "disabled")

	// define four gates
	gateTestBaz1 := codegate.New("Baz1")
	gateTestBaz2 := codegate.New("Baz2")
	gateTestBaz3 := codegate.New("Baz3")
	gateTestBaz4 := codegate.New("Baz4")

	require.False(t, gateTestBaz1.Enabled(), "Baz1 should be disabled")
	require.True(t, gateTestBaz2.Enabled(), "Baz2 should be enabled")
	require.False(t, gateTestBaz3.Enabled(), "Baz3 should be disabled")
	require.True(t, gateTestBaz4.Enabled(), "Baz4 should be enabled")

	require.Contains(t, codegate.DisabledGates(), "Baz1")
	require.NotContains(t, codegate.DisabledGates(), "Baz2")
	require.Contains(t, codegate.DisabledGates(), "Baz3")
	require.NotContains(t, codegate.DisabledGates(), "Baz4")
}
