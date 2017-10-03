#include <stdio.h>
#include <stdlib.h>
#include <time.h>

double calcPifloat();
double gd;

#define NITER 40
#define ROUNDS 10000

int main(void) {
	double ld;
	double sum = 0;
	for (int i = 0; i < NITER; ++i) {
		clock_t start_clock = clock();
		for (int j = 0; j < ROUNDS; ++j) {
			ld = calcPifloat();
		}
		sum += (double)(clock() - start_clock)/(double)(CLOCKS_PER_SEC);
	}
	gd = ld;
	printf("average: %f\n", sum / NITER);
	return 0;
}

double calcPifloat() {
	double lasts = 0.0;
	double t     = 3.0;
	double s     = 3.0;
	double n     = 1.0;
	double na    = 0.0;
	double d     = 0.0;
	double da    = 24.0;
	while (s != lasts) {
		lasts = s;
		n += na;
		na += 8;
		d += da;
		da += 32;
		t = (t * n) / d;
		s = t;
	}
	return s;
}
