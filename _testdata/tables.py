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
    "A": "abs",
    "cfd": "convert-to-string",
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
        x = random.randint(0, 6)
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

def write_line(out, prec, op, mode, r, x, y = None):
    if x is not None:
        if y is not None:
            str = "d{}{} {} {} {} -> {}\n".format(prec, op, mode, x, y, r)
        else:
            str = "d{}{} {} {} -> {}\n".format(prec, op, mode, x, r)
    else:
        raise ValueError("bad args")
    out.write(str)

def perform_op(op):
    x = rand_dec()
    y = None # possibly unused

    # Binary
    if op == "*":
        y = rand_dec()
        r = x * y
    elif op == "+":
        y = rand_dec()
        r = x + y
    elif op == "-":
        y = rand_dec()
        r = x - y
    elif op == "/":
        y = rand_dec()
        r = x / y
    elif op == "qC":
        y = rand_dec()
        r = x.compare(y)
    elif op == "quant":
        y = rand_dec(True)
        with decimal.localcontext() as c:
            c.prec = decimal.MAX_PREC
            r = c.quantize(x, y)
        y = -y.as_tuple().exponent

    # Unary
    elif op == "A":
        r = x.copy_abs()
    elif op == "cfd":
        r = str(x)
    else:
        raise ValueError("bad op")
    return (r, x, y)

# set N higher for local testing.
N = 5000

def make_tables():
    for op, name in ops.items():
        with gzip.open("{}-tables.gz".format(name), "wt") as f:
            for i in range(1, N):
                mode = random.choice(list(modes.keys()));

                ctx = decimal.getcontext()
                ctx.rounding = modes[mode]
                ctx.prec = random.randint(1, 5000)
               
                t = perform_op(op)
                write_line(f, ctx.prec, op, mode, *t[:3])


make_tables()
