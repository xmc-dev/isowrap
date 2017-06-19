#include <stdio.h>

int main(void) {
	int i, j;
	int a[8];
	for (i = 0; i < 100000000; ++i) {
		for (j = 0; j < 100000000; ++j) {
			a[i & 7] += a[j & 7];
		}
	}
	printf("%d %d", a[0], a[3]);
	return 0;
}
