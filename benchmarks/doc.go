// Package benchmarks implements benchmarking across decimal packages.
//
// General Notes
//
//    1. Times are measured in seconds, unless otherwise noted
//    2. Measured on a MacBook Pro, 2.9 GHz Intel Core i5, 8 GB 2133 MHz LPDDR3
//    3. Some benchmarks are adapted from www.bytereef.org/mpdecimal/benchmarks.html
//
// While the benchmarks aim to be as fair as possible, the packages have
// different feature sets and objects.
//
// For example, go-inf/inf[8] boasts the fastest overall runtime of any package,
// but does not fully implement the GDA specification: it lacks contexts,
// non-finite values (NaN, Inf, or ±0), conditions, etc.
//
// Further, programs like cockroachdb/apd[2] sacrifice speed to ensure strict
// compliance with the GDA spec.
//
// In general, package sthat cannot fully complete a challenge will be unranked.
// For example, the ``float64'' type cannot provide 19 or more digits of precision,
// so it's unranked in the Pi test. Similarly so with apmckinlay/dnum[3].
//
// Pi
//
// Go programs are tested with Go 1.12.2.
//
// Java benchmarks are separated into two categories, warm and cold. Warm
// benchmarks are run 10,000 times prior to benchmarking.
//
//    |    Program (version)                 | 9 digits | 19 digits | 38 digits | 100 digits | average |
//    |--------------------------------------|----------|-----------|-----------|------------|---------|
//    | go-inf/inf[8]                        | 0.0912   | 0.201     | 0.426     | 1.280      | 0.499   |
//    | ericlagergren/decimal[1] (mode Go)   | 0.0477   | 0.234     | 0.522     | 1.690      | 0.623   |
//    | Python decimal[5] (Python 3.7.3)     | 0.143    | 0.322     | 0.657     | 2.325      | 0.862   |
//    | JDK BigDecimal[4] (Java 1.8, warm)   | 0.0570   | 0.196     | 0.721     | 3.060      | 1.008   |
//    | ericlagergren/decimal[1] (mode GDA)  | 0.0510   | 0.333     | 0.922     | 3.550      | 1.214   |
//    | JDK BigDecimal[4] (Java 1.8, cold)   | 0.302    | 0.854     | 1.308     | 3.100      | 1.391   |
//    | shopspring/decimal[7] decimal        | 0.289    | 0.700     | 1.490     | 4.120      | 1.649   |
//    | cockroachdb/apd[2]                   | 0.564    | 2.410     | 10.000    | 78.300     | 22.818  |
//    | Python decimal[6] (Python 2.7.10)    | 31.81    | 74.502    | 161.71    | 460.00     | 182.00  |
//    | float64                              | 0.0557   | -         | -         | -          | -       |
//    | double (C LLVM 10.0.1 -O3)           | 0.0589   | -         | -         | -          | -       |
//    | apmckinlay/dnum[3]                   | 0.0456   | -         | -         | -          | -       |
//    | float (Python 2.7.10, 3.7.3)         | 0.0923   | -         | -         | -          | -       |
//
// Mandelbrot
//
// Go programs are tested with Go 1.9.?.
//
//    |    Program (version)                 | 9 digits | 16 digits | 19 digits | 34 digits | 38 digits | average |
//    |--------------------------------------|----------|-----------|-----------|-----------|-----------|---------|
//    | ericlagergren/decimal[1] (mode GDA)  | 2.73     | 9.07      | 14.54     | 24.95     | 25.09     | 15.27   |
//    | ericlagergren/decimal[1] (mode Go)   | 2.73     | 9.70      | 15.02     | 26.13     | 26.62     | 16.04   |
//    | float64                              | 0.0034   | -         | -         | -         | -         | -       |
//
// References
//
// Further information and references can be found at the links below.
//
//    [1]: github.com/ericlagergren/decimal
//    [2]: github.com/cockroachdb/apd
//    [3]: github.com/apmckinlay/gsuneido/util/dnum
//    [4]: docs.oracle.com/javase/8/docs/api/java/math/BigDecimal.html
//    [5]: docs.python.org/3.6/library/decimal.html
//    [6]: docs.python.org/2/library/decimal.html
//    [7]: github.com/shopspring/decimal
//    [8]: github.com/go-inf/inf
//
package decimal_benchmarks
