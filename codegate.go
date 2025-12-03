package codegate

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
)

// Gate is a code gate, allowing code to be selectively enabled or disabled.
type Gate struct {
	name    string
	enabled bool
}

// EnvVarPrefix is the prefix for environment variables used to disable
// code gates. This prefix is defined in an exported variable so that it
// may be changed at compile time with build flags:
//
//	go build -ldflags "-X 'github.com/singlestore-labs/codegate.EnvVarPrefix=MY_PREFIX_'"
//
// This variable should not be changed at runtime including in init functions,
// because it is used by the New function which may be called in static
// initializers.
var (
	EnvVarPrefix = "DISABLE_CODE_"
)

const (
	// Deprecated: use DISABLE_CODE_ prefix instead
	envVarPrefix2 = "DISABLE_S2CODE_"
	nameMaxLength = 100
)

var (
	// gate names must be valid environment variable names
	validName     = regexp.MustCompile("^[A-Za-z][A-Za-z0-9_]*$")
	usedNames     = map[string]struct{}{}
	disabledGates []string
	gateLock      sync.Mutex
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
	if _, found := usedNames[name]; found {
		panic(fmt.Errorf(`code gate name (%s) is already in use. Code gate names must be unique`, name))
	}
	usedNames[name] = struct{}{}
	_, disabled := os.LookupEnv(EnvVarPrefix + name)
	if !disabled {
		_, disabled = os.LookupEnv(envVarPrefix2 + name)
	}
	return Gate{
		name:    name,
		enabled: !disabled,
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

// DisabledGates returns the names of all currently disabled code gates. The
// list is loaded from the environment variables and includes all variables
// prefixed with the code gate prefix regardless of whether a gate has been
// created for that name.
func DisabledGates() []string {
	gateLock.Lock()
	defer gateLock.Unlock()
	if disabledGates == nil {
		disabledGates = []string{}
		// Get all disabled code gates from the environment variables.
		for _, env := range os.Environ() {
			envName, _, _ := strings.Cut(env, "=")
			if strings.HasPrefix(envName, EnvVarPrefix) {
				disabledGates = append(disabledGates, strings.TrimPrefix(envName, EnvVarPrefix))
			} else if strings.HasPrefix(envName, envVarPrefix2) {
				disabledGates = append(disabledGates, strings.TrimPrefix(envName, envVarPrefix2))
			}
		}
	}
	return disabledGates
}

func resetDisabledGates() {
	gateLock.Lock()
	defer gateLock.Unlock()
	disabledGates = nil
}
