#include "distance.h"
#include <immintrin.h>
#include <math.h>  // Ensure math library is included

// Function to calculate the Euclidean distance using AVX
float euclidean_distance_avx(const float* a, const float* b, size_t length) {
    __m256 sum = _mm256_setzero_ps();
    size_t i;
    
    for (i = 0; i + 8 <= length; i += 8) {
        __m256 vec_a = _mm256_loadu_ps(a + i);
        __m256 vec_b = _mm256_loadu_ps(b + i);
        __m256 diff = _mm256_sub_ps(vec_a, vec_b);
        __m256 sqr = _mm256_mul_ps(diff, diff);
        sum = _mm256_add_ps(sum, sqr);
    }

    float result[8];
    _mm256_storeu_ps(result, sum);
    
    float distance = 0.0f;
    for (int j = 0; j < 8; j++) {
        distance += result[j];
    }

    // Handle remaining elements
    for (; i < length; i++) {
        float diff = a[i] - b[i];
        distance += diff * diff;
    }

    return sqrtf(distance);
}
