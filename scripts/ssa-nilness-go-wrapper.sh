#!/bin/sh
set -eu
if [ "${1:-}" = test ]; then
  shift
  exec "$NILNESS_REAL_GO" test \
    -run '^Test(StringerNilness$|NonNil|NilIndex|UnsupportedArguments|IndexCalls|BuildPanic|QueryPanic|NoReturn)' "$@"
fi
exec "$NILNESS_REAL_GO" "$@"
