package math_test

import (
	"testing"

	"github.com/ericlagergren/decimal/internal/test"
)

func TestLog(t *testing.T) { test.Log.Test(t) }

/*
Benchmarks from "Handbook of Continued Fractions for Special Functions."

alg1 - 2.4.1
alg2 - 2.4.4
alg3 - 2.4.7

With a small value for x ("42.2") and a precision of 16, the number of iterations
was:

alg1 - 210
alg2 - 75
alg3 - 150

Code for the algorithm behind the benchmarks:
https://gist.github.com/ericlagergren/cc95be6530aec21e7f91e2204173fd4f

BenchmarkLog_alg1_9-4     	     500	   3209008 ns/op
BenchmarkLog_alg1_19-4    	     300	   5114247 ns/op
BenchmarkLog_alg1_38-4    	     200	  12034146 ns/op
BenchmarkLog_alg1_500-4   	       2	 535323033 ns/op
BenchmarkLog_alg2_9-4     	    3000	    478031 ns/op
BenchmarkLog_alg2_19-4    	    1000	   1954844 ns/op
BenchmarkLog_alg2_38-4    	     300	   4615867 ns/op
BenchmarkLog_alg2_500-4   	       5	 238076617 ns/op
BenchmarkLog_alg3_9-4     	    2000	   1043696 ns/op
BenchmarkLog_alg3_19-4    	     300	   4317666 ns/op
BenchmarkLog_alg3_38-4    	     200	   8040413 ns/op
BenchmarkLog_alg3_500-4   	       2	 550735383 ns/op
*/
