import java.math.BigDecimal;
import java.math.MathContext;
import java.util.function.Function;

public class JavaBigDecimal {
    private static final BigDecimal TWO = new BigDecimal(2);
    private static final BigDecimal EIGHT = new BigDecimal(8);
    private static final BigDecimal THIRTY_TWO = new BigDecimal(32);

    private static final int[] PI_PRECS = {9, 19, 38, 100};
    private static final int PI_ITERS = 10000;

    private static final int[] MBROT_PRECS = {9, 16, 19, 34, 38, 100};
    private static final int MBROT_ITERS = 1;

    public static void main(String... args) {
        BigDecimal r = BigDecimal.ZERO;
        int[] precs;
        int iters;
        Function<MathContext, BigDecimal> fn;

        for (final String arg : args) {
            switch (arg) {
                case "pi":
                    fn = (MathContext mc) -> pi(mc);
                    precs = PI_PRECS;
                    iters = PI_ITERS;
                    break;
                case "mbrot":
                    fn = (MathContext mc) -> mbrot(mc);
                    precs = MBROT_PRECS;
                    iters = MBROT_ITERS;
                    break;
                default:
                    throw new IllegalArgumentException("bad argument " + args);
            }

            for (int i = 0; i < 10000; ++i) {
                r = fn.apply(new MathContext(precs[i % precs.length]));
            }

            double sum = 0.0;

            for (final int p : precs) {
                long start = System.currentTimeMillis();
                for (int i = 0; i < iters; i++) {
                    r = fn.apply(new MathContext(p));
                }
                final double end = (System.currentTimeMillis() - start) / 1000.0;
                sum += end;
                System.out.printf("%d: time: %fs\n", p, end);
                //System.out.println(r.toString());
            }
            System.out.printf("average: %f\n", sum / precs.length);
        }
    }

    private static BigDecimal pi(MathContext mc) {
        BigDecimal lasts = new BigDecimal(0);
        BigDecimal t = new BigDecimal(3);
        BigDecimal s = new BigDecimal(3);
        BigDecimal n = new BigDecimal(1);
        BigDecimal na = new BigDecimal(0);
        BigDecimal d = new BigDecimal(0);
        BigDecimal da = new BigDecimal(24);

        while (s.compareTo(lasts) != 0) {
            lasts = s;
            n = n.add(na, mc);
            na = na.add(EIGHT, mc);
            d = d.add(da, mc);
            da = da.add(THIRTY_TWO, mc);
            t = t.multiply(n, mc);
            t = t.divide(d, mc);
            s = s.add(t, mc);
        }
        return s;
    }

    private static BigDecimal mbrot(MathContext mc) {
        BigDecimal x0 = new BigDecimal("0.222");
        BigDecimal y0 = new BigDecimal("0.333");
        BigDecimal x = x0;
        BigDecimal y = y0;
        BigDecimal xx = x.multiply(x, mc);
        BigDecimal yy = y.multiply(y, mc);
        for (int i = 0; i < 10000000; ++i) {
            y = x.multiply(y, mc);
            y = y.multiply(TWO, mc);
            y = y.add(y0, mc);
            x = xx.subtract(yy, mc);
            x = x.add(x0, mc);
            xx = x.multiply(x, mc);
            yy = y.multiply(y, mc);
        }
        return x;
    }
}

