# codegate

Code gates allow bug fixes, system-wide features, and other code changes to be selectively disabled. Code gates support rapid
rollback of new code in production when an unanticipated bug or side effect is discovered. Code gates are enabled by default
but may be disabled without redelivering software through a CI/CD pipeline.

## Wrapping Code in a Gate

Code gates are intended to be short term constructs. Once the gated code is validated and stable at production, the code gate
should be removed to simplify future development. In most cases gates should be removed in a matter of weeks, not months or
years. A TODO comment should be included with each gate declaration as a reminder. In most cases a corresponding ticket should
be created to track remove of the gate. Title or camel case, i.e., Go style naming, is recommended. A standard prefix for each
code domain (e.g., RBAC for RBAC related behaviors) is suggested to avoid name conflicts. The code gate is normally declared
in a global variable using a static initializer call:

```go
// TODO: TICKET-1743 remove this code gate once gated code is stable at production
var gateRBACExtraCleanup = codegate.New("RBACExtraCleanup")


func doSomething() {
    if gateRBACDeleteOrphanedGrants.Enabled() {
        // execute new code
        deleteOrphanedGrants(ctx)
    } else {
        // execute old code replaced by new code above, if any
    }
}
```

The New() function panics if the name is blank, invalid, or has already been used. Calling codegate.New() outside a static
initializer is not supported and will frequently cause a runtime panic. Code gates are typically local to a single package
but, if necessary, the gate global variable can be exported and used in other packages. The gate still must be created only
once.

## Controlling the Gate

Code gates are controlled by environment variables. Code gates are created in the enabled state unless an environment
variable named `DISABLE_<codegate-name>` is found in the process environment. The value of the variable is ignored;
any value, including blank, is sufficient to disable the code gate. The normal use case for a code gate is to disable or
revert some code by defining the code gate environment variable and *restarting* the service or application.

Code gates are immutable. The enable/disable state is determined at creation time and will not be affected by modifying
the corresponding environment variable after creation.

