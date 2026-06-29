#!/usr/bin/env python3
"""Strip control characters (U+0000–U+001F except tab/newline/cr) from stdin."""
import sys
import re

data = sys.stdin.buffer.read().decode("utf-8", errors="replace")
cleaned = re.sub(r"[\x00-\x08\x0b\x0c\x0e-\x1f]", "", data)
sys.stdout.write(cleaned)
