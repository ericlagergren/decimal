import java.math.BigDecimal;
import java.math.MathContext;

public class JavaBigDecimal {
    private static final BigDecimal EIGHT = new BigDecimal(8);
    private static final BigDecimal THIRTY_TWO = new BigDecimal(32);

    public static void main(String... args) {
        BigDecimal r = new BigDecimal(0);

        final int[] prec = {9, 19, 38, 100};
        for (int i = 0; i < 10000; ++i) {
            r = pi(new MathContext(prec[i % prec.length]));
        }

        double sum = 0.0;

        for (final int p : prec) {
            long start = System.currentTimeMillis();
            for (int i = 0; i < 10000; i++) {
                r = pi(new MathContext(p));
            }
            final double end = (System.currentTimeMillis() - start) / 1000.0;
            sum += end;
            System.out.printf("time: %fs\n", end);
            System.out.println(r.toString());
        }
        System.out.printf("average: %f\n", sum / prec.length);
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
}

