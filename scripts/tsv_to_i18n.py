#!/usr/bin/env python

import sys
import json
import codecs

if len(sys.argv) < 2:
    sys.stderr.write("Usage: %s filename\n" % sys.argv[0])
    sys.exit(1)

f = codecs.open(sys.argv[1], "r", "utf-8")
header = f.readline().strip().split("\t")
lines = (l.strip("\n") for l in f)
rows = (l.split("\t") for l in lines)
tuples_rows = (zip([""]+header, row) for row in rows)
tuples_rows2 = ([(tr[0][1], t) for t in tr[1:]] for tr in tuples_rows)
tuples = (t for tr in tuples_rows2 for t in tr)
i18n = reduce(lambda d,c: d.setdefault(c[1][0], {}).update({c[0]: c[1][1].replace("\\n", "\n").replace("\\t", "\t")}) or d, tuples, {})

print json.dumps(i18n, ensure_ascii=False, sort_keys=True, indent=4, separators=(",", ": ")).encode("utf-8")
