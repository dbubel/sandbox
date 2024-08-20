#include <immintrin.h>
#include <stdio.h>

void simd_matrix_add(float *a, float *b, float *c, int rows, int cols) {
    int total_elements = rows * cols;
    int simd_width = 8; // AVX processes 8 floats at a time
    int i;

    // Process elements in chunks of 8
    for (i = 0; i <= total_elements - simd_width; i += simd_width) {
        __m256 va = _mm256_loadu_ps(&a[i]);
        __m256 vb = _mm256_loadu_ps(&b[i]);
        __m256 vc = _mm256_add_ps(va, vb);
        _mm256_storeu_ps(&c[i], vc);
    }

    // Process remaining elements
    for (; i < total_elements; i++) {
        c[i] = a[i] + b[i];
    }
}

int main() {
    int rows = 4, cols = 8; // Example dimensions
    float a[32] = {
        1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0,
        9.0, 10.0, 11.0, 12.0, 13.0, 14.0, 15.0, 16.0,
        17.0, 18.0, 19.0, 20.0, 21.0, 22.0, 23.0, 24.0,
        25.0, 26.0, 27.0, 28.0, 29.0, 30.0, 31.0, 32.0
    };
    float b[32] = {
        32.0, 31.0, 30.0, 29.0, 28.0, 27.0, 26.0, 25.0,
        24.0, 23.0, 22.0, 21.0, 20.0, 19.0, 18.0, 17.0,
        16.0, 15.0, 14.0, 13.0, 12.0, 11.0, 10.0, 9.0,
        8.0, 7.0, 6.0, 5.0, 4.0, 3.0, 2.0, 1.0
    };
    float c[32];

    simd_matrix_add(a, b, c, rows, cols);

    for (int i = 0; i < rows; i++) {
        for (int j = 0; j < cols; j++) {
            printf("%f ", c[i * cols + j]);
        }
        printf("\n");
    }

    return 0;
}
