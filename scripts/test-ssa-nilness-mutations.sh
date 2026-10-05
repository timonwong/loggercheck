#!/bin/sh
set -eu

# Mutate a disposable copy so the deliberately failing repros remain intact.
repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
output=${1:?usage: test-ssa-nilness-mutations.sh /absolute/path/results.json}
case "$output" in /*) ;; *) echo 'Output path must be absolute' >&2; exit 2 ;; esac
real_go=$(command -v go)
gremlins=$(command -v gremlins)
task_dir=$(mktemp -d "${TMPDIR:-/tmp}/loggercheck-mutations.XXXXXX")
trap 'rm -rf "$task_dir"' EXIT HUP INT TERM
mkdir -p "$task_dir/repo" "$task_dir/bin"
cp -R "$repo_dir/internal" "$repo_dir/testdata" "$task_dir/repo/"
cp "$repo_dir/go.mod" "$repo_dir/go.sum" "$repo_dir/loggercheck.go" \
  "$repo_dir/loggercheck_test.go" "$repo_dir/options.go" "$repo_dir/staticrules.go" "$task_dir/repo/"
if [ -f "$task_dir/repo/testdata/src/a/nilness/adversarial_repro.go" ]; then
  rm "$task_dir/repo/testdata/src/a/nilness/adversarial_repro.go"
fi
cp "$repo_dir/scripts/ssa-nilness-go-wrapper.sh" "$task_dir/bin/go"
chmod +x "$task_dir/bin/go"
export NILNESS_REAL_GO="$real_go"
export PATH="$task_dir/bin:$PATH"
cd "$task_dir/repo"
# Integration mode is needed because the analyzer fixtures cover another package.
# The wrapper restricts every test invocation to the nilness tests.
"$gremlins" unleash --integration --coverpkg ./internal/nilness/... \
  --exclude-files '^internal/(checkers|rules|sets|stringutil)/|^(loggercheck|options|staticrules).*\.go$|_test\.go$|^testdata/' \
  --invert-logical --workers 4 --timeout-coefficient 10 --output "$output"
