#include <stdio.h>
int add(int x, int y) {
  if (y != 0) {
    return add(x^y,(x&y) << 1);
  }
  return x;
}
int main() {
  printf("%d",add(5,5));
}
