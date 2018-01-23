#include <stdio.h>
#include <string.h>
#include <stdlib.h>

#define MULT (5 << 20)

int main(void)
{
    int i;
    static char x[MULT];
    x[0] = x[1] = 1;
    while (1) {
        for (i = 2; i < MULT; ++i) {
            x[i] = x[i - 1] + x[i - 2];
        }
    }
    return 0;
}
