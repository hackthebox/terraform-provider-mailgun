#!/bin/bash
# Copyright (c) Hack The Box
# SPDX-License-Identifier: MPL-2.0

set -o errexit
set -o nounset

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__parent="$(dirname "$__dir")"
__root="$(dirname "$__parent")"

CHANGELOG_FILE_NAME="CHANGELOG.md"
CHANGELOG_TMP_FILE_NAME="CHANGELOG.tmp"
TARGET_SHA=$(git rev-parse HEAD)

# Get the last release tag, or use initial commit if no tags exist
PREVIOUS_RELEASE_TAG=$(git describe --abbrev=0 --match='v*.*.*' --tags 2>/dev/null || echo "")

if [ -z "$PREVIOUS_RELEASE_TAG" ]; then
    echo "No previous release tag found, using initial commit"
    PREVIOUS_RELEASE_SHA=$(git rev-list --max-parents=0 HEAD)
else
    PREVIOUS_RELEASE_SHA=$(git rev-list -n 1 $PREVIOUS_RELEASE_TAG)
fi

if [ "$TARGET_SHA" == "$PREVIOUS_RELEASE_SHA" ]; then
    echo "Nothing to do"
    exit 0
fi

echo "Generating changelog from $PREVIOUS_RELEASE_SHA to $TARGET_SHA"

CHANGELOG=$(changelog-build -this-release "$TARGET_SHA" \
                      -last-release "$PREVIOUS_RELEASE_SHA" \
                      -git-dir "$__root" \
                      -entries-dir "$__root/.changelog" \
                      -changelog-template "$__dir/changelog.tmpl" \
                      -note-template "$__dir/release-note.tmpl" \
                      -local-fs)

if [ -z "$CHANGELOG" ]; then
    echo "No changelog generated."
    exit 0
fi

echo "Generated changelog:"
echo "$CHANGELOG"
