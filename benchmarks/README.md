## Benchmarks for various decimal programs.

For info on specific benchmarks, visit
http://www.bytereef.org/mpdecimal/benchmarks.html

### Pi

Times are in seconds.

|    Program                  | 9 digits |  19 digits  | 38 digits | 100 digits | average |
|-----------------------------|----------|-------------|-----------|------------|---------|
| BigDecimal (Java 1.8, warm) | 0.049    | 0.19        | 0.6       | 3.29       | 1.05    |
| decimal (Go 1.9)            | 0.039    | 0.53        | 1.30      | 4.45       | 1.57    |
| decimal (Python 3.6.2)      | 0.27     | 0.58        | 1.32      | 4.52       | 1.67    |
| BigDecimal (Java 1.8)       | 0.29     | 0.96        | 1.79      | 3.99       | 1.76    |
| apd (Go 1.9)                | 0.52     | 2.21        | 8.98      | 68.43      | 20.03   |
| decimal (Python 2.7.10)     | 12.93    | 28.91       | 64.96     | 192.58     | 74.84   |
| float64 (Go 1.9)            | 0.057    | -           | -         | -          | -       |
| double (C LLVM 9.0.0 -O3)   | 0.057    | -           | -         | -          | -       |
| float (Python 2.7.10)       | 0.59     | -           | -         | -          | -       |
