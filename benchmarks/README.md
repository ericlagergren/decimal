## Benchmarks for various decimal programs.

- Times are measured in seconds, unless otherwise noted
- Measured on a MacBook Pro, 2.9 GHz Intel Core i5, 8 GB 2133 MHz LPDDR3
- Some benchmarks are adapted from www.bytereef.org/mpdecimal/benchmarks.html

## Pi

|    Program (version)                      | 9 digits |  19 digits  | 38 digits | 100 digits | average |
|-------------------------------------------|----------|-------------|-----------|------------|---------|
| [go-inf/inf][8] (Go 1.9)                       | 0.10     | 0.23        | 0.53      | 1.43       | 0.572   |
| [JDK BigDecimal][4] (Java 1.8, warm)           | 0.049    | 0.19        | 0.6       | 3.29       | 1.05    |
| [ericlagergren/decimal][1] (Go 1.9, mode Go)   | 0.034    | 0.40        | 1.00      | 3.57       | 1.25    |
| [Python decimal][5] (Python 3.6.2)             | 0.27     | 0.58        | 1.32      | 4.52       | 1.67    |
| [ericlagergren/decimal][1] (Go 1.9, mode GDA)  | 0.048    | 0.55        | 1.46      | 4.91       | 1.74    |
| [JDK BigDecimal][4] (Java 1.8)                 | 0.29     | 0.96        | 1.79      | 3.99       | 1.76    |
| [shopspring/decimal][7] decimal (Go 1.9)       | 0.38     | 0.94        | 1.95      | 5.26       | 2.13    |
| [cockroachdb/apd][2] (Go 1.9)                  | 0.52     | 2.14        | 9.01      | 71.62      | 20.81   |
| [Python decimal][6] (Python 2.7.10)            | 12.93    | 28.91       | 64.96     | 192.58     | 74.84   |
| float64 (Go 1.9)                          | 0.057    | -           | -         | -          | -       |
| double (C LLVM 9.0.0 -O3)                 | 0.057    | -           | -         | -          | -       |
| [apmckinlay/dnum][3] (Go 1.9)                  | 0.091    | -           | -         | -          | -       |
| float (Python 2.7.10)                     | 0.59     | -           | -         | -          | -       |

[1]: https://github.com/ericlagergren/decimal
[2]: https://github.com/cockroachdb/apd
[3]: https://github.com/apmckinlay/gsuneido/util/dnum
[4]: https://docs.oracle.com/javase/8/docs/api/java/math/BigDecimal.html
[5]: https://docs.python.org/3.6/library/decimal.html
[6]: https://docs.python.org/2/library/decimal.html
[7]: https://github.com/shopspring/decimal
[8]: https://github.com/go-inf/inf
