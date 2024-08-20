#include <stdio.h>
#include <stdlib.h>
#include <time.h>

#define MAX_LINE_LENGTH 512

// Function to generate a random printable ASCII character
char randomChar() {
    return ' ' + (rand() % 95); // Printable ASCII characters range from 32 to 126
}

// Function to generate a random line of text up to a specified maximum length
void generateRandomLine(char *line, int maxLength) {
    int lineLength = rand() % (maxLength + 1); // Random length between 0 and maxLength
    for (int i = 0; i < lineLength; i++) {
        line[i] = randomChar();
    }
    line[lineLength] = '\0'; // Null-terminate the string
}

int main(int argc, char *argv[]) {
    if (argc != 3) {
        fprintf(stderr, "Usage: %s <filename> <number_of_lines>\n", argv[0]);
        return 1;
    }

    const char *filename = argv[1];
    int numLines = atoi(argv[2]);

    FILE *file = fopen(filename, "w");
    if (file == NULL) {
        perror("Failed to open file");
        return 1;
    }

    srand(time(NULL)); // Seed the random number generator

    char line[MAX_LINE_LENGTH + 1]; // Buffer to hold the generated line
    for (int i = 0; i < numLines; i++) {
        generateRandomLine(line, MAX_LINE_LENGTH);
        fprintf(file, "%s\n", line);
    }

    fclose(file);
    printf("Successfully wrote %d lines to %s\n", numLines, filename);

    return 0;
}
