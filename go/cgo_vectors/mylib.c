// mylib.c
#include "mylib.h"

// Function to add two vectors
void add_vectors(float* a, float* b, float* result, int length) {
    for (int i = 0; i < length; i++) {
        result[i] = a[i] + b[i];
    }
}
