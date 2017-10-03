#!/usr/bin/env python

import time
from math import log, ceil
import decimal

Decimal = decimal.Decimal

def pi_native(prec = None):
    """native float"""
    lasts, t, s, n, na, d, da = 0, 3.0, 3, 1, 0, 0, 24
    while s != lasts:
        lasts = s
        n, na = n+na, na+8
        d, da = d+da, da+32
        t = (t * n) / d
        s += t
    return s

def pi_decimal(prec):
    """Decimal"""
    decimal.getcontext().prec = prec
    lasts, t, s, n, na, d, da = Decimal(0), Decimal(3), Decimal(3), Decimal(1), Decimal(0), Decimal(0), Decimal(24)
    while s != lasts:
        lasts = s
        n, na = n+na, na+8
        d, da = d+da, da+32
        t = (t * n) / d
        s += t
    return s

d = {
    "native":  pi_native,
    "decimal": pi_decimal,
}

for name in ["native", "decimal"]:
    print(name)
    sum = 0.0
    for prec in [9, 19, 38, 100]:
        func = d[name]
        start = time.time()
        for i in range(10000):
            x = func(prec)
        x = time.time()-start
        sum += x
        print("%fs\n" % x)
    print("avg: %f\n" % (sum / 4.0))
