#!/usr/bin/env bash

# Copyright The XSTS-SH Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -eo pipefail

ROOT_DIR="$(git rev-parse --show-toplevel)"

ensure_sponge() {
  if command -v sponge >/dev/null 2>&1; then
    return
  fi

  echo "sponge not found; attempting to install moreutils..."

  # Use sudo if available and not already running as root.
  SUDO=""
  if [ "$(id -u)" -ne 0 ] && command -v sudo >/dev/null 2>&1; then
    SUDO="sudo"
  fi

  if command -v apt-get >/dev/null 2>&1; then
    $SUDO apt-get update -y
    $SUDO apt-get install -y moreutils
  elif command -v yum >/dev/null 2>&1; then
    $SUDO yum install -y moreutils
  elif command -v dnf >/dev/null 2>&1; then
    $SUDO dnf install -y moreutils
  else
    echo "Error: sponge (moreutils) is required but could not be installed automatically." >&2
    echo "Please install the 'moreutils' package manually." >&2
    exit 1
  fi

  if ! command -v sponge >/dev/null 2>&1; then
    echo "Error: sponge (moreutils) installation appears to have failed." >&2
    exit 1
  fi
}

ensure_sponge

GO_FILES=$(find "$ROOT_DIR" -not -path "/vendor/*" -type f -name '*.go')
PY_FILES=$(find "$ROOT_DIR" -not -path "/venv/*" -type f -name '*.py')

GO_TPL="$ROOT_DIR/hack/boilerplate.go.txt"
PY_TPL="$ROOT_DIR/hack/boilerplate.py.txt"

for file in $GO_FILES; do
  if ! grep -q "Copyright The XSTS-SH Authors" "$file"; then
    (cat "$GO_TPL" && echo && cat "$file") | sponge "$file"
  fi
done

for file in $PY_FILES; do
  if ! grep -q "Copyright The XSTS-SH Authors" "$file"; then
    (cat "$PY_TPL" && echo && cat "$file") | sponge "$file"
  fi
done

echo "Update Copyright Done"