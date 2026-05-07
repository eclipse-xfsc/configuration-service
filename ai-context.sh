#!/usr/bin/env sh

set -eu

exec go -C ./.ai run . --repo-root .. "$@"
