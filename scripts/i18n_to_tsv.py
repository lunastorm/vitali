#!/usr/bin/env python

import sys
import json

if len(sys.argv) < 2:
    sys.stderr.write("Usage: %s filename\n" % sys.argv[0])
    sys.exit(1)

with open(sys.argv[1]) as f:
    i18n = json.load(f)

langs = sorted([lang for lang in i18n])
header = "\t"+"\t".join(langs)
print header

labels = sorted(reduce(lambda r,lang: r.update(i18n[lang]) or r, langs, set()))
rows = ("%s\t%s" % (label, "\t".join([i18n[lang][label].replace("\n", "\\n").replace("\t", "\\t") if label in i18n[lang] else "" for lang in langs])) for label in labels)

print "\n".join(rows).encode("utf-8")
