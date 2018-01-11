#!/usr/bin/env python3

from decimal import *
import gzip
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
    "%": "remainder",
    "Nu": "next-plus",
    "Nd": "next-minus",

    # Custom
    "rat": "convert-to-rat",
    "sign": "sign",
    "signbit": "signbit",
    "exp": "exponential-function",
    "log": "natural-logarithm",
    "log10": "common-logarithm",
    "pow": "power",
    "//": "integer-division",
    "norm": "reduction",
    "rtie": "round-to-integral-exact",
    "shift": "shift",
}

modes = {
    "=0": ROUND_HALF_EVEN,
    "=^": ROUND_HALF_UP,
    "0": ROUND_DOWN,
    "<": ROUND_FLOOR,
    ">": ROUND_CEILING,
    # ROUND_HALF_DOWN,
    "^": ROUND_UP,
    # ROUND_05UP,
}


def rand_bool():
    return random.randint(0, 1) % 2 == 0


def make_dec(nbits=5000):
    r = random.randint(0, 50)
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

    return Decimal(fs)


DEC_TEN = Decimal(10)


def rand_dec(quant=None, nbits=None):
    if quant is None:
        quant = False
    if nbits is None:
        nbits = 5000
    with localcontext() as ctx:
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
    if isinstance(x, Decimal):
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
    x = None
    y = None  # possibly unused
    u = None  # possibly unused

    try:
        # Binary
        if op == "*":
            x = rand_dec()
            y = rand_dec()
            r = x * y
        elif op == "+":
            x = rand_dec()
            y = rand_dec()
            r = x + y
        elif op == "-":
            x = rand_dec()
            y = rand_dec()
            r = x - y
        elif op == "/":
            x = rand_dec()
            y = rand_dec()
            r = x / y
        elif op == "//":
            x = rand_dec()
            y = rand_dec()
            r = x // y
        elif op == "%":
            x = rand_dec()
            y = rand_dec()
            r = x % y
        elif op == "qC":
            x = rand_dec()
            y = rand_dec()
            r = x.compare(y)
        elif op == "quant":
            x = rand_dec()
            y = rand_dec(True)
            r = x.quantize(y)
            y = -y.as_tuple().exponent
        elif op == "pow":
            getcontext().prec += 11
            getcontext().prec //= 10
            x = rand_dec(nbits=64)
            y = rand_dec(nbits=64)
            #u = rand_dec(nbits=64)
            # The error of Python's decimal power method is < 1 ULP + t, where
            # t <= 0.1 ULP, but usually < 0.01 ULP.
            getcontext().prec += 1
            r = getcontext().power(x, y, u)
            getcontext().prec -= 1
            r = +r
        elif op == "shift":
            x = rand_dec()
            y = Decimal(random.randint(-getcontext().prec, getcontext().prec))
            r = x.shift(y)

        # Unary
        elif op == "A":
            x = rand_dec()
            r = getcontext().abs(x)
        elif op == "cfd":
            x = rand_dec()
            r = str(x)
        elif op == "rat":
            x = rand_dec()
            while True:
                try:
                    x, y = x.as_integer_ratio()
                    if y == 1:
                        r = +Decimal(x)
                    else:
                        r = Decimal(x) / Decimal(y)
                    break
                except Exception:  # ValueError if nan, etc.
                    x = rand_dec()
        elif op == "sign":
            x = rand_dec()
            if x < 0:
                r = -1
            elif x > 0:
                r = +1
            else:
                r = 0
        elif op == "signbit":
            x = rand_dec()
            r = x.is_signed()
        elif op == "~":
            x = rand_dec()
            r = -x
        elif op == "exp":
            if getcontext().prec >= 10:
                getcontext().prec //= 10
            x = rand_dec(nbits=100)
            r = x.exp()
        elif op == "log":
            getcontext().prec += 11
            getcontext().prec //= 10
            x = rand_dec(nbits=128)
            r = x.ln()
        elif op == "L":
            getcontext().prec += 11
            getcontext().prec //= 10
            x = rand_dec(nbits=128)
            r = x.logb()
        elif op == "log10":
            getcontext().prec += 11
            getcontext().prec //= 10
            x = rand_dec(nbits=128)
            r = x.log10()
        elif op == "?":
            x = rand_dec()
            r = x.number_class()
        elif op == "V":
            x = rand_dec()
            r = x.sqrt()
        elif op == "norm":
            x = rand_dec()
            r = x.normalize()
        elif op == "rtie":
            x = rand_dec()
            r = x.to_integral_exact()
        elif op == "Nu":
            x = rand_dec()
            r = x.next_plus()
        elif op == "Nd":
            x = rand_dec()
            r = x.next_minus()

        # Ternary
        elif op == "*-":
            x = rand_dec()
            y = rand_dec()
            u = rand_dec()
            r = x.fma(y, u)
        else:
            raise ValueError("bad op {}".format(op))
    except Exception as e:
        raise e

    return (r, x, y, u)


traps = {
    Clamped: "c",
    DivisionByZero: "z",
    Inexact: "x",
    InvalidOperation: "i",
    Overflow: "o",
    Rounded: "r",
    Subnormal: "s",
    Underflow: "u",
    FloatOperation: "***",
}


def rand_traps():
    t = {}
    s = ""
    for key, val in traps.items():
        b = key != FloatOperation and rand_bool()
        if b:
            s += val
        t[key] = int(b)
    return (t, s)


# set N higher for local testing.
N = 2500


def make_tables(items):
    for op, name in items:
        with gzip.open("{}-tables.gz".format(name), "wt") as f:
            for i in range(1, N):
                mode = random.choice(list(modes.keys()))
                # t, ts = rand_traps()

                ctx = getcontext()
                ctx.Emax = MAX_EMAX
                ctx.Emin = MIN_EMIN
                ctx.rounding = modes[mode]
                ctx.prec = random.randint(1, 5000)
                ctx.clear_traps()
                ctx.clear_flags()

                r, x, y, u = perform_op(op)

                conds = ""
                for key, value in ctx.flags.items():
                    if value == 1 and key != FloatOperation:
                        conds += traps[key]
                write_line(f, ctx.prec, op, mode, r, x, y, u, conds)


items = ops.items()
if len(sys.argv) > 1:
    arg = sys.argv[1]
    items = [(arg, ops[arg])]
make_tables(items)
