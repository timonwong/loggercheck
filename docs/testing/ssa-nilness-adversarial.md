# SSA nilness adversarial testing

Baseline: `c2be25e` (PR #124). Test seam: the public loggercheck analyzer,
through `analysistest` and source-level `// want` diagnostics. Implementation
changes are out of scope. A missing diagnostic for a reachable nil pointer is
a soundness failure; a diagnostic for a safe pointer is conservative.

## Failure inventory (written before fixtures)

1. **Incorrect no-return contracts.** klog's replaceable `OsExit` can return;
   Fatal, Exit and FlushAndExit must not delete a reachable nil edge. Package
   identity must survive vendoring without mistaking a user `log`, `os`,
   `runtime`, `testing` or klog package for the trusted library. Prefix matches
   can trust unrelated Fatal/Panic/Exit names. Receiver identity matters for
   testing methods. Function values and interface calls need conservative
   treatment. Panic/recover must not make a deferred logger unreachable.
2. **Incorrect condition polarity or dominance.** OR/AND guards, boolean
   comparisons, expression/value/type switches, selects, goto, labeled
   break/continue, zero-iteration loops and shadowed variables must not transfer
   a guard to an unsafe path or different SSA value.
3. **Stale value facts.** Reassignment to nil/unknown, closure writes,
   pointer-to-pointer aliases, field reloads and synchronized goroutine writes
   invalidate facts about earlier loads. Facts about immutable SSA operands
   must not become facts about mutable storage.
4. **Wrong comparison operand.** Comparing two pointers, typed-nil interfaces,
   `any(p) != nil`, or reflection does not supply a pointer/nil branch fact.
5. **Unsound phi recursion.** Nested merges, nil backedges, self/mutual cycles
   and ChangeType must check every feasible incoming value. Assuming a cycle
   non-nil must still reject any reachable nil entry or backedge.
6. **Non-panicking or misplaced dereference evidence.** Deferred/closure
   dereferences, recovered panics, non-dominating dereferences, a different
   operand, nil map/slice operations and pointer-receiver methods cannot prove
   the logged pointer non-nil. Array-pointer `len` need not dereference;
   value-receiver calls do dereference. Calls before a later dereference remain
   unsafe. Cross-function referrers cannot establish dominance.
7. **Generics and new lowering.** Type-parameter candidates may escape the
   concrete-pointer checker entirely. Guarded and unguarded instances,
   range-over-func/yield closures, integer ranges, min/max/clear must retain
   diagnostics or conservatively fall back if SSA construction fails.
8. **Wrong call/value mapping.** `Lparen` is a token offset, not a displayed
   line: same-line calls, repeated `//line` locations, generated files,
   defer/go, chained methods and external test packages must map to their own
   operand and program point. Ambiguous positions must stay unknown. Variadic
   slot arithmetic must distinguish unsafe and safe values in one call.
9. **Missing function indexing.** Package-variable closures, multiple init
   functions, nested closures, method expressions and generic instantiations
   must be indexed or conservatively reported, never confused with another call.

Each group has safe controls as well as unsafe inputs where applicable. The
fixtures use real vendored klog/logr; arbitrary named-package impostors are
separate fixture dependencies. Runtime witnesses exercise reachable nil paths
without invoking a nil-unsafe String method themselves.

## Results

There are **two SSA soundness bugs**, plus a **pre-existing generic candidate
gap** outside SSA. Five logging sites intentionally lack expected diagnostics.
The implementation files are unchanged.

### Confirmed false negatives

| Fixtures | Minimal logging path | Root cause on baseline | Suggested repair |
| --- | --- | --- | --- |
| `adversarialReturningExitHook`, `adversarialReturningFatalHook` | Replace `klog.OsExit` with `func(int){}`; `if p == nil { klog.FlushAndExit(..., 1) }`; log `p` | `internal/nilness/nilness.go:170-172` trusts functions whose exit contract is mutable, deleting the nil predecessor | Conservatively remove klog's hook-backed functions from noReturn. Proving a default hook requires interprocedural/global mutation knowledge. |
| `adversarialVendoredLog`, `adversarialVendoredOS` | A returning `x/vendor/log.Fatal` or `x/vendor/os.Exit` in the nil branch; log `p` | `internal/nilness/nilness.go:153-164` strips the vendor prefix before trusting stdlib identity | Match stdlib paths exactly before any devendoring; restrict library names/receivers to known contracts. Devendoring alone does not establish a no-return contract. |
| `adversarialGenericPointer`, `adversarialGenericInstantiation` | `P interface{ *U; String() string }`, instantiated with `U=NamespacedName`, called with nil | `internal/checkers/checker.go:51-54` rejects `*types.TypeParam` before querying SSA | Include concrete pointer terms/instances with a proven value-receiver Stringer element; keep unproven nilness conservative. The same pointer-only gate is present in `640df0f` (v0.12.1); do not count this gap as an SSA regression. |

klog v2.70.1 is the real fixture dependency, not a stub:
`exit.go:38-51` exports `OsExit` and calls it in FlushAndExit; `klog.go:950`
calls it on the Fatal path. The runtime witness confirms both calls return
with a nil value at the logging point. It also confirms the generic typed-nil
value is non-nil when boxed and panics when its String method is called.
See `nilness-runtime-witness.json`.

For vendoring, GOPATH `go list -deps x` confirms bindings to `x/vendor/log`
and `x/vendor/os`, not the standard packages. The returning stubs are tracked
explicitly despite the repository's global vendor ignore rule.

### Conservative false positives

The passing fixtures preserve the actual diagnostic with `// want`; these
annotations record accepted conservative behavior, not ideal suppression.

| Fixture | Safe situation | Impact / missing proof |
| --- | --- | --- |
| `adversarialBooleanEquality` | After `(p != nil) == false` returns | Boolean-expression normalization is not recognized. |
| `adversarialTypedNilInterface` | `any(p)` differs from a typed-nil `*NamespacedName` interface | The same dynamic type makes this a non-nil guard, but interface comparisons are not modeled. |
| `adversarialReflection` | `!reflect.ValueOf(p).IsNil()` | Reflection needs call semantics not modeled here. |
| `adversarialMethodExpression` | `(*NamespacedName).String(p)` returned | Dereference occurs inside the generated receiver wrapper, outside local SSA. |
| `adversarialFunctionValue` | `f := klog.Fatal; f(...)` with the default terminating hook | An indirect callee has no noReturn contract; suppression would need the same hook caveat as direct Fatal. |
| `adversarialTestingInterface` | `testing.TB.Fatal` | Interface calls have no static callee. |
| `adversarialPanicRecover` (normal path) | A panic exits the function; the normal continuation has non-nil `p` | Capturing `p` makes the guard and log separate heap-cell loads. The deferred log remains correctly unsafe. |

Previously recorded conservative cases remain: field reloads, closure-captured
loads, error-result relationships, local no-return helpers, and unreachable
calls removed from SSA. These produce extra warnings, not lost warnings.

### Mutation testing

Tool: Gremlins **v0.6.0**, Go **1.26.3**, default mutators plus
`--invert-logical`, limited to `internal/nilness/nilness.go`.

| Run | Killed | Lived | Not covered | Timeout / not viable | Score |
| --- | ---: | ---: | ---: | ---: | ---: |
| Initial adversarial fixtures | 59 | 8 | 0 | 0 / 0 | 88.06% |
| After gap fixtures | 64 | 3 | 0 | 0 / 0 | 95.52% |

The score is `KILLED / (KILLED + LIVED)` without discounting equivalent mutants.
Mutant coverage is 100% in both runs. The JSON files include every mutant's
type, line, column and status. This is the selected Gremlins operator set,
not an exhaustive proof against every possible implementation mutation.

All initial survivors and their classification:

| Line:column / operator | Mutation | Classification and evidence |
| --- | --- | --- |
| 153:49 / CONDITIONALS_BOUNDARY | `i >= 0` to `i > 0` | Equivalent for legal import paths: an occurrence at zero would require an import path beginning `/vendor/`, which Go rejects. Remains LIVED. |
| 160:20 / INVERT_LOGICAL | os gate `&&` to `||` | Test gap. `adversarialReturningOSFunction` prevents treating `os.Getpid` as no-return. Now KILLED. |
| 162:20 / INVERT_LOGICAL | runtime gate `&&` to `||` | Test gap. `adversarialReturningRuntimeFunction` prevents treating `runtime.Gosched` as no-return. Now KILLED. |
| 288:71 / INVERT_LOGICAL | right-hand operand/nil conjunction to disjunction | Test gap. `adversarialRightPointerComparison` logs the right operand of a comparison to another pointer. Now KILLED. |
| 288:66 / CONDITIONALS_NEGATION | `binop.Y == v` to `!=` | Test gap. `adversarialReversedNilGuard` exercises `nil != p`. Now KILLED. |
| 302:12 / INVERT_LOGICAL | nil-constant gate `&&` to `||` | Test gap. `adversarialDerefThenPointerComparison` needs a failed nonconstant comparison check to continue to valid dereference proof; the mutant panics and conservatively reports. Now KILLED. |
| 335:32 / INVERT_LOGICAL | unary dereference gate `&&` to `||` | Equivalent for pointer candidates in well-typed SSA: a pointer used by UnOp can only be the X operand of `*`, not a receive, boolean negation or arithmetic operand. Referrer enumeration already guarantees that X is v. Remains LIVED. |
| 342:16 / INVERT_LOGICAL | pointer IndexAddr gate `&&` to `||` | Equivalent for pointer candidates in well-typed SSA: a pointer cannot be an integer index, so its IndexAddr referrer uses it as X, necessarily an array pointer. Remains LIVED. |

The last two equivalence claims depend on the documented pointer-only query
contract and valid Go SSA. They do not claim equivalence for arbitrary fabricated
SSA or nonpointer API misuse.

Reproduce the mutation run after vendoring fixture dependencies:

```sh
cd testdata/src/a && go mod vendor && cd ../../..
go install github.com/go-gremlins/gremlins/cmd/gremlins@v0.6.0
sh scripts/test-ssa-nilness-mutations.sh "$PWD/docs/testing/nilness-mutations-final.json"
```

The script uses a disposable copy, removes only `adversarial_repro.go` there,
and restricts every Gremlins test invocation to the existing nilness unit tests
and `TestStringerNilness`. The deliberately failing vendored-name test is also
outside that subset. This prevents an already failing baseline from falsely
marking every mutant KILLED. It does not run the full E2E suite per mutant.

Runtime witness (run in the nested fixture module):

```sh
NILNESS_WITNESS_ARTIFACT="$PWD/docs/testing/nilness-runtime-witness.json" \
  sh -c 'cd testdata/src/a && go test -count=1 ./runtimewitness'
```

Go has no source macros; the generated same-line fixture and repeated `//line`
locations cover the corresponding call-position hazard. `positions_generated.go`
intentionally retains two calls on one physical line, so do not reformat it.

### Full-suite verification

Both final runs used `go test -count=1 ./...` on **timon-m3mac**, with each
toolchain's bin directory first in PATH so go/packages also uses that version.

| Toolchain | Full-suite result | Artifact |
| --- | --- | --- |
| Go 1.26.3 darwin/arm64 | FAIL: only the five missing diagnostics above | `go1.26.3-full.log` |
| Go 1.27.1 darwin/arm64 | FAIL: the same five missing diagnostics | `go1.27.1-full.log` |

The passing fixture set and existing nilness unit tests were also checked
independently before adding the failing commits. Runtime witnesses passed on
Go 1.26.3. Full suites ran once per requested toolchain at the end; development
and mutation runs used the nilness subset.

The 71 named adversarial/external fixture functions cover the inventory above.
This is an adversarial regression corpus, not a claim of complete soundness.
The branch deliberately retains correct failing expectations for every
confirmed false negative. No implementation fix is included.
