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

def rand_dec():
    return decimal.Decimal("{}{}".format(
        "+" if random.randint(0, 1) % 2 == 0 else "-",
        random.uniform(0, sys.float_info.min)))

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
        

