## Benchmarks contains benchmarks for various decimal programs. Not all
benchmarks have been collected by me. For info on specific benchmarks, visit
http://www.bytereef.org/mpdecimal/benchmarks.html

### Pi

Times in seconds.

|    Program    | 9 digits |  19 digits  | 38 digits | 100 digits | average |
|---------------|----------|-------------|-----------|------------|---------|
| gmpy          | 0.52     | 0.52        | 1.07      | 3.57       | 1.42    |
| cdecimal-nt   | 0.24     | 0.55        | 1.21      | 4.08       | 1.52    |
| decimal Go    | 0.039    | 0.53        | 1.30      | 4.45       | 1.57    |
| cdecimal      | 0.27     | 0.58        | 1.32      | 4.52       | 1.67    |
| Java          | 0.38     | 0.73        | 1.25      | 5.35       | 1.92    |
| apd Go        | 0.52     | 2.21        | 8.98      | 68.43      | 20.03   |
| decimal       | 17.61    | 42.75       | -         | -          | 30.18   |
| float64 Go    | 0.057    | -           | -         | -          | N/A     |
| Python float  | 0.12     | -           | -         | -          | N/A     |
