#!/usr/bin/env python

import gzip
import decimal
import random
import sys

ops = {
    "*": "multiplication",
    "+": "addition",
    "-": "subtraction",
    "/": "division",
}

def make_dec():
    sign = "+" if random.randint(0, 1) % 2 == 0 else "-"
    return decimal.Decimal("{}{}".format(sign, random.uniform(0, sys.float_info.min)))

def rand_dec():
    q = random.randint(0, 4)
    d = make_dec()
    if q == 1:
        d *= make_dec()
    elif q == 2:
        d /= make_dec()
    elif q == 3:
        d -= make_dec()
    else: # q == 4
        d += make_dec()
    return d

for op, name in ops.items():
    with gzip.open("{}-tables.gzip".format(name), "wt") as f:
        for i in range(1, 10000):
            prec = random.randint(1, 5000)
            decimal.getcontext().prec = prec
            x = rand_dec()
            y = rand_dec()
            if op == "*":
                r = x * y
            elif op == "+":
                r = x + y
            elif op == "-":
                r = x - y
            else:
                r = x / y
            f.write("d{}{} =0 {} {} -> {}\n".format(prec, op, x, y, r))
        

