package codegate

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
)

// Gate is a code gate, allowing code to be selectively enabled or disabled.
// Gate currently contains only a Name field, but may grow in the future to
// include other metadata about the gate.
type Gate struct {
	name    string
	enabled bool
}

// EnvVarPrefix is the prefix for environment variables used to disable
// code gates. This prefix is defined in an exported variable so that it
// may be changed at compile time with build flags:
//
//	go build -ldflags "-X 'github.com/singlestore-labs/codegate.EnvVarPrefix=MY_PREFIX_'"
var EnvVarPrefix = "DISABLE_CODE_"

const nameMaxLength = 100

var (
	validName = regexp.MustCompile("^[A-Za-z][A-Za-z0-9_]*$")
	usedNames = map[string]struct{}{}
	gateLock  sync.Mutex
)

// New creates a code gate. Code gate names must be globally unique and should
// be defined in static initializers. For example,
//
//	var gateRBACDeleteOrphanedGrants = codegate.New("RBACDeleteOrphanedGrants")
//
// The gate name must start with an alpha character and contain only alphanumeric
// or underbar characters. Title casing (i.e., Go style naming) and standard prefix
// for each code domain (e.g., "RBAC" for RBAC related behaviors) is recommended.
// New panics if the name is missing, invalid, or is a duplicate.
func New(name string) Gate {
	if !validName.MatchString(name) || len(name) > nameMaxLength {
		panic(fmt.Errorf(`code gate name (%s) is invalid. Code gate names must begin with an alpha, contain only alphanumerics or underbars, and be no more than %d characters in length`,
			name, nameMaxLength))
	}
	gateLock.Lock()
	defer gateLock.Unlock()
	if _, ok := usedNames[name]; ok {
		panic(fmt.Errorf(`code gate name (%s) is already in use. Code gate names must be unique`, name))
	}
	usedNames[name] = struct{}{}
	_, ok := os.LookupEnv(EnvVarPrefix + name)
	return Gate{
		name:    name,
		enabled: !ok,
	}
}

// Enabled returns true unless the code gate been disabled. Code gates
// control system-wide features, bug fixes, or other behavioral changes
// that may need to be disabled due to unforeseen failures or side effects.
//
// Example usage:
//
//	if gateRBACDeleteOrphanedGrants.Enabled() {
//		deleteOrphanedGrants(ctx)
//	} else {
//		// execute old code replaced by new code above, if any
//	}
//
// NOTE: Code gates are currently disabled by defining the DISABLE_<gate-name>
// environment variable for the process(es) implementing the code gate. Any Go
// process may implement code gates. This implementation may change in the
// future. Runtime code outside this package should not have any dependencies on
// the environment variable implementation.
func (gate Gate) Enabled() bool {
	return gate.enabled
}

// Name returns the code gate name.
func (gate Gate) Name() string {
	return gate.name
}

// String returns a string representation of the code gate and its state.
func (gate Gate) String() string {
	label := fmt.Sprintf("code gate %s", gate.name)
	if gate.enabled {
		return label + " (enabled)"
	}
	return label + " (disabled)"
}

// DisabledGates returns the names of all disabled code gates based on
// current environment variables.
func DisabledGates() []string {
	disabledGates := []string{}
	for _, env := range os.Environ() {
		envName, _, _ := strings.Cut(env, "=")
		if strings.HasPrefix(envName, EnvVarPrefix) {
			disabledGates = append(disabledGates, strings.TrimPrefix(envName, EnvVarPrefix))
		}
	}
	return disabledGates
}
