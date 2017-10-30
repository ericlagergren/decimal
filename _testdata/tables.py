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
    "qC": "comparison",
    "quant": "quantize",
}

modes = {
    "=0": decimal.ROUND_HALF_EVEN,
    "=^": decimal.ROUND_HALF_UP,
    "0": decimal.ROUND_DOWN, 
    "<": decimal.ROUND_FLOOR,
    ">": decimal.ROUND_CEILING,
    # decimal.ROUND_HALF_DOWN,
    "^": decimal.ROUND_UP,
    # decimal.ROUND_05UP,
}

def rand_bool():
    return random.randint(0, 1) % 2 == 0

def make_dec():
    sign = "+" if rand_bool() else "-"
    return decimal.Decimal("{}{}".format(sign, random.uniform(1, sys.float_info.max)))

DEC_TEN = decimal.Decimal(10)

def rand_dec(quant = False):
    if quant:
        x = random.randint(0, 250)
        if rand_bool():
            d = DEC_TEN ** x
        else:
            d = DEC_TEN ** -x
    else:
        q = random.randint(0, 4)
        d = make_dec()
        if q == 1:
            d *= make_dec()
        elif q == 2:
            d /= make_dec()
        elif q == 3:
            d -= make_dec()
        elif q == 4:
            d += make_dec()
        # else: == 0
    return d

# set N higher for local testing.
N = 5000

for op, name in ops.items():
    with gzip.open("{}-tables.gzip".format(name), "wt") as f:
        for i in range(1, N):
            prec = random.randint(1, 5000)
            decimal.getcontext().prec = prec

            mode = random.choice(list(modes.keys()));
            decimal.getcontext().rounding = modes[mode]

            x = rand_dec()
            y = rand_dec(op == "quant")
            if op == "*":
                r = x * y
            elif op == "+":
                r = x + y
            elif op == "-":
                r = x - y
            elif op == "/":
                r = x / y
            elif op == "qC":
                r = x.compare(y)
            elif op == "quant":
                decimal.getcontext().prec = decimal.MAX_PREC
                r = x.quantize(y)
                y = y.as_tuple().exponent
            else:
                raise ValueError("bad op")
            f.write("d{}{} {} {} {} -> {}\n".format(prec, op, mode, x, y, r))
