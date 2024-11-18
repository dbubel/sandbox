#include <stdio.h>
#include <string.h>

void print_array(int* arr, int size) {
    for(int i = 0; i < size; i++) {
        printf("%d ", arr[i]);
    }
    printf("\n");
}

int main() {
    int array1[5] = {1, 2, 3, 4, 5};
    int array2[5] = {10, 20, 30, 40, 50};
    
    printf("Memory layout:\n");
    printf("array1 at: %p\n", (void*)array1);
    printf("array2 at: %p\n", (void*)array2);
    
    printf("\nProper access of array1: ");
    print_array(array1, 5);
    
    printf("Reading beyond array1 bounds: ");
    // This will read memory beyond array1, potentially including array2
    print_array(array1, 10);
    
    return 0;
}
