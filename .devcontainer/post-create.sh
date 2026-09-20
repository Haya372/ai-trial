#!/bin/bash
set -e

mise install
mise exec -- pnpm install

# bubblewrap runs inside a container and cannot mount a fresh /proc,
# so enable the weaker nested sandbox mode for this environment only.
python3 - <<'EOF'
import json, os
p = '.claude/settings.local.json'
c = json.load(open(p)) if os.path.exists(p) else {}
c.setdefault('sandbox', {})['enableWeakerNestedSandbox'] = True
with open(p, 'w') as f:
    json.dump(c, f, indent=2)
    f.write('\n')
EOF
