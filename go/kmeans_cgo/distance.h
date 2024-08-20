#ifndef DISTANCE_H
#define DISTANCE_H

#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

// Function to calculate the Euclidean distance using AVX
float euclidean_distance_avx(const float* a, const float* b, size_t length);

#ifdef __cplusplus
}
#endif

#endif // DISTANCE_H
