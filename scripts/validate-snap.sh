#!/usr/bin/env bash
# validate-snap.sh — Validate snap/snapcraft.yaml for correctness and Snap Store compliance.
# Requirements: 1.3, 1.5, 7.1, 7.3
#
# Checks performed:
#   1. YAML is syntactically valid
#   2. Required metadata fields are present
#   3. Summary is ≤ 79 characters
#   4. License is a recognised SPDX identifier
#   5. Confinement is "strict"
#
# Exit codes:
#   0 — all checks passed
#   1 — one or more checks failed

set -euo pipefail

YAML_FILE="${1:-snap/snapcraft.yaml}"

if [ ! -f "$YAML_FILE" ]; then
  echo "ERROR: $YAML_FILE not found."
  exit 1
fi

# Run all validation inside a single Python 3 script.
# We attempt to use PyYAML for full syntax validation; if it is not installed
# we fall back to a simple line-based parser that covers the top-level scalar
# fields needed for the checks.

python3 - "$YAML_FILE" << 'PYEOF'
import sys
import os

yaml_path = sys.argv[1]

# ---------------------------------------------------------------------------
# 1. Parse the YAML file
# ---------------------------------------------------------------------------
parsed = {}
yaml_available = False

try:
    import yaml  # PyYAML
    yaml_available = True
except ImportError:
    pass

with open(yaml_path, "r") as fh:
    raw = fh.read()

if yaml_available:
    try:
        parsed = yaml.safe_load(raw)
        if not isinstance(parsed, dict):
            print("ERROR: YAML root is not a mapping.")
            sys.exit(1)
        print("OK   : YAML syntax is valid (verified with PyYAML).")
    except yaml.YAMLError as exc:
        print(f"ERROR: YAML syntax error — {exc}")
        sys.exit(1)
else:
    # Fallback: basic line-based parser for top-level scalar fields.
    # This does NOT validate full YAML syntax but is good enough for the
    # metadata checks when PyYAML is unavailable.
    print("WARN : PyYAML not installed — using basic line parser (limited syntax check).")
    for line in raw.splitlines():
        stripped = line.rstrip()
        # Skip blank lines, comments, and indented lines (not top-level).
        if not stripped or stripped.startswith("#") or stripped[0] in (" ", "\t"):
            continue
        if ":" in stripped:
            key, _, value = stripped.partition(":")
            key = key.strip()
            value = value.strip().strip("'\"")
            if value:  # only capture scalar values
                parsed[key] = value

# ---------------------------------------------------------------------------
# 2. Validate required fields
# ---------------------------------------------------------------------------
errors = []

REQUIRED_FIELDS = [
    "name",
    "version",
    "summary",
    "description",
    "license",
    "grade",
    "base",
    "confinement",
]

for field in REQUIRED_FIELDS:
    if field not in parsed:
        errors.append(f"Missing required field: {field}")
    else:
        print(f"OK   : Field '{field}' is present.")

# ---------------------------------------------------------------------------
# 3. Summary length ≤ 79 characters
# ---------------------------------------------------------------------------
summary = parsed.get("summary", "")
if isinstance(summary, str):
    length = len(summary)
    if length > 79:
        errors.append(f"Summary is {length} characters (max 79): \"{summary}\"")
    else:
        print(f"OK   : Summary is {length} characters (≤ 79).")
elif summary is not None:
    errors.append(f"Summary is not a string: {type(summary).__name__}")

# ---------------------------------------------------------------------------
# 4. License is a recognised SPDX identifier
# ---------------------------------------------------------------------------
# A minimal set of common SPDX identifiers. Extend as needed.
KNOWN_SPDX = {
    "AGPL-3.0",
    "AGPL-3.0-only",
    "AGPL-3.0-or-later",
    "Apache-2.0",
    "BSD-2-Clause",
    "BSD-3-Clause",
    "GPL-2.0",
    "GPL-2.0-only",
    "GPL-2.0-or-later",
    "GPL-3.0",
    "GPL-3.0-only",
    "GPL-3.0-or-later",
    "ISC",
    "LGPL-2.1",
    "LGPL-3.0",
    "MIT",
    "MPL-2.0",
    "Proprietary",
    "Unlicense",
}

license_val = parsed.get("license", "")
if isinstance(license_val, str):
    if license_val in KNOWN_SPDX:
        print(f"OK   : License '{license_val}' is a valid SPDX identifier.")
    else:
        errors.append(
            f"License '{license_val}' is not in the known SPDX list: "
            f"{', '.join(sorted(KNOWN_SPDX))}"
        )
elif license_val is not None:
    errors.append(f"License is not a string: {type(license_val).__name__}")

# ---------------------------------------------------------------------------
# 5. Confinement must be "strict"
# ---------------------------------------------------------------------------
confinement = parsed.get("confinement", "")
if confinement == "strict":
    print("OK   : Confinement is 'strict'.")
else:
    errors.append(
        f"Confinement must be 'strict' for Snap Store compliance, got '{confinement}'."
    )

# ---------------------------------------------------------------------------
# Report results
# ---------------------------------------------------------------------------
if errors:
    print()
    print("VALIDATION FAILED:")
    for err in errors:
        print(f"  ✗ {err}")
    sys.exit(1)
else:
    print()
    print("All checks passed ✓")
    sys.exit(0)
PYEOF
