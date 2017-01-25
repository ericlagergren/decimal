#include <iostream>
#include <math.h>
#include <boost/math/tools/fraction.hpp>

template <class T>
struct cosine_fraction
{
    private:
        T a, b;
        T m;
        T z;
    public:
        cosine_fraction(T z) 
            : a(0), b(0), z(z), m(-1)
        {}

        typedef std::pair<T,T> result_type;

        std::pair<T,T> operator()()
        {
            m++;
            if (m == 0) {
                return std::make_pair(0, 1);
            };
            a = (z*z) / (m * (((T)4*m) + -2));
            b = 1 - a;
            std::cout << a << " / " << b << " = " << a/b << std::endl;
            return std::make_pair(a, b);
        }
};

template <class T>
T cosine(T a)
{
    cosine_fraction<T> fract(a);
    T v =  boost::math::tools::continued_fraction_b(
            fract,
            std::numeric_limits<T>::epsilon()
    );
    return 1 / v;
}

int main(void) {
    auto z = 4.0;
    std::cout << cos(z) << std::endl;
    std::cout << cosine(z) << std::endl;
}
