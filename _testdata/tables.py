#!/usr/bin/env python3

import gzip
import decimal
import random
import sys
import math

ops = {
    "*": "multiplication",
    "+": "addition",
    "-": "subtraction",
    "/": "division",
    "qC": "comparison",
    "quant": "quantization",
    "A": "absolute-value",
    "cfd": "convert-to-string",
    "~": "negation",
    "*-": "fused-multiply-add",
    "L": "base-b-logarithm",
    "?": "class",
    "V": "square-root",

    # Custom
    "rat": "convert-to-rat",
    "sign": "sign",
    "signbit": "signbit",
    "exp": "exponential-function",
    "log": "natural-logarithm",
    "log10": "common-logarithm",
    "pow": "power",
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


def make_dec(nbits=54321):
    r = random.randint(0, 25)
    if r == 0:
        f = math.nan
    elif r == 1:
        f = math.inf
    else:
        f = random.getrandbits(random.randint(1, nbits))

    if rand_bool():
        if r > 1:
            f = -f
        else:
            f = math.copysign(f, -1)

    fs = "{}".format(f)
    if rand_bool():
        l = list(fs)
        if len(l) == 1:
            l.insert(0, '.')
        else:
            l[random.randint(0, len(l) - 1)] = '.'
        fs = "".join(l)

    return decimal.Decimal(fs)


DEC_TEN = decimal.Decimal(10)


def rand_dec(quant=False, nbits=54321):
    with decimal.localcontext() as ctx:
        ctx.clear_traps()

        if quant:
            x = random.randint(0, 750)
            if rand_bool():
                d = DEC_TEN ** x
            else:
                d = DEC_TEN ** -x
        else:
            d = make_dec(nbits)
            for _ in range(random.randint(0, 3)):
                q = random.randint(0, 4)
                if q == 1:
                    d *= make_dec(nbits)
                elif q == 2:
                    d /= make_dec(nbits)
                elif q == 3:
                    d -= make_dec(nbits)
                elif q == 4:
                    d += make_dec(nbits)
                # else: == 0
    return d


def conv(x):
    if x is None:
        return None
    if isinstance(x, decimal.Decimal):
        if x.is_infinite():
            return '-Inf' if x.is_signed() else 'Inf'
        else:
            return x
    return x


def write_line(out, prec, op, mode, r, x, y=None, u=None, flags=None):
    if x is None:
        raise ValueError("bad args")

    x = conv(x)
    y = conv(y)
    u = conv(u)
    r = conv(r)
    if y is not None:
        if u is not None:
            str = "d{}{} {} {} {} {} -> {} {}\n".format(
                prec, op, mode, x, y, u, r, flags)
        else:
            str = "d{}{} {} {} {} -> {} {}\n".format(
                prec, op, mode, x, y, r, flags)
    else:
        str = "d{}{} {} {} -> {} {}\n".format(
            prec, op, mode, x, r, flags)
    out.write(str)


def perform_op(op):
    r = None
    x = rand_dec(nbits=64 if op == "pow" else None)
    y = None  # possibly unused
    u = None  # possibly unused

    try:
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
            r = x.quantize(y)
            y = -y.as_tuple().exponent
        elif op == "pow":
            y = rand_dec(nbits=64)
            #u = rand_dec(nbits=64)
            r = decimal.getcontext().power(x, y, u)

        # Unary
        elif op == "A":
            r = decimal.getcontext().abs(x)
        elif op == "cfd":
            r = str(x)
        elif op == "rat":
            while True:
                try:
                    x, y = x.as_integer_ratio()
                    if y == 1:
                        r = +decimal.Decimal(x)
                    else:
                        r = decimal.Decimal(x) / decimal.Decimal(y)
                    break
                except Exception:  # ValueError if nan, etc.
                    x = rand_dec()
        elif op == "sign":
            if x < 0:
                r = -1
            elif x > 0:
                r = +1
            else:
                r = 0
        elif op == "signbit":
            r = x.is_signed()
        elif op == "~":
            r = -x
        elif op == "exp":
            r = x.exp()
        elif op == "log":
            r = x.ln()
        elif op == "L":
            r = x.logb()
        elif op == "log10":
            r = x.log10()
        elif op == "?":
            r = x.number_class()
        elif op == "V":
            r = x.sqrt()

        # Ternary
        elif op == "*-":
            y = rand_dec()
            u = rand_dec()
            r = x.fma(y, u)
        else:
            raise ValueError("bad op {}".format(op))
    except Exception as e:
        raise e

    return (r, x, y, u)


traps = {
    decimal.Clamped: "c",
    decimal.DivisionByZero: "z",
    decimal.Inexact: "x",
    decimal.InvalidOperation: "i",
    decimal.Overflow: "o",
    decimal.Rounded: "r",
    decimal.Subnormal: "s",
    decimal.Underflow: "u",
    decimal.FloatOperation: "***",
}


def rand_traps():
    t = {}
    s = ""
    for key, val in traps.items():
        b = key != decimal.FloatOperation and rand_bool()
        if b:
            s += val
        t[key] = int(b)
    return (t, s)


# set N higher for local testing.
N = 100


def make_tables(items):
    for op, name in items:
        with gzip.open("{}-tables.gz".format(name), "wt") as f:
            for i in range(1, N):
                mode = random.choice(list(modes.keys()))
                # t, ts = rand_traps()

                ctx = decimal.getcontext()
                ctx.Emax = decimal.MAX_EMAX
                ctx.Emin = decimal.MIN_EMIN
                ctx.rounding = modes[mode]
                ctx.prec = random.randint(1, 5000)
                ctx.clear_traps()
                ctx.clear_flags()

                r, x, y, u = perform_op(op)

                conds = ""
                for key, value in ctx.flags.items():
                    if value == 1 and key != decimal.FloatOperation:
                        conds += traps[key]
                write_line(f, ctx.prec, op, mode, r, x, y, u, conds)


items = ops.items()
if len(sys.argv) > 1:
    arg = sys.argv[1]
    items = [(arg, ops[arg])]
make_tables(items)
