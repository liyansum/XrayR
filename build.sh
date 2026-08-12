#!/bin/sh
set -eu

task_go=${GO:-go}
task_binary=${BINARY:-XrayR}
task_goamd64=${GOAMD64:-v3}
task_ldflags=${LDFLAGS:--s -w}

CGO_ENABLED=0 GOAMD64="$task_goamd64" "$task_go" build \
  -trimpath \
  -ldflags="$task_ldflags" \
  -o "$task_binary" \
  .
