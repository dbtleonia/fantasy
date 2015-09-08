#!/usr/bin/python

import itertools
import sys

if len(sys.argv) != 2:
    sys.exit('Usage: %s <num-rounds>' % sys.argv[0])

for r in xrange(0, int(sys.argv[1])):
    for p in itertools.combinations_with_replacement('DKQRTW', r):
        print ''.join(p)
