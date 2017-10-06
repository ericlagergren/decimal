## Benchmarks for various decimal programs.

- Times are measured in seconds unless otherwise noted
- Measured on a MacBook Pro, 2.9 GHz Intel Core i5, 8 GB 2133 MHz LPDDR3
- Some benchmarks are adapted from
www.bytereef.org/mpdecimal/benchmarks.html

## Pi
|    Program (version)                  | 9 digits |  19 digits  | 38 digits | 100 digits | average |
|-----------------------------|----------|-------------|-----------|------------|---------|
| BigDecimal<sup>[4]</sup> (Java 1.8, warm) | 0.049    | 0.19        | 0.6       | 3.29       | 1.05    |
| decimal<sup>[1]</sup> (Go 1.9)            | 0.039    | 0.53        | 1.30      | 4.45       | 1.57    |
| decimal<sup>[5]</sup> (Python 3.6.2)      | 0.27     | 0.58        | 1.32      | 4.52       | 1.67    |
| BigDecimal (Java 1.8)       | 0.29     | 0.96        | 1.79      | 3.99       | 1.76    |
| apd<sup>[2]</sup> (Go 1.9)                | 0.52     | 2.21        | 8.98      | 68.43      | 20.03   |
| decimal<sup>[6]</sup> (Python 2.7.10)     | 12.93    | 28.91       | 64.96     | 192.58     | 74.84   |
| float64 (Go 1.9)            | 0.057    | -           | -         | -          | -       |
| double (C LLVM 9.0.0 -O3)   | 0.057    | -           | -         | -          | -       |
| float (Python 2.7.10)       | 0.59     | -           | -         | -          | -       |
| dnum<sup>[3]</sup> (Go 1.9)               | 0.95     | -           | -         | -          | -       |

[1]: https://github.com/ericlagergren/decimal
[2]: https://github.com/cockroachdb/apd
[3]: https://github.com/apmckinlay/gsuneido/util/dnum
[4]: https://docs.oracle.com/javase/8/docs/api/java/math/BigDecimal.html
[5]: https://docs.python.org/3.6/library/decimal.html
[6]: https://docs.python.org/2/library/decimal.html
