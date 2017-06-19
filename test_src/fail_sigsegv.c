#include <stdio.h>

int main() {
	int *p = 0xDEADBEEF;
	*p = 2;
	printf("%d\n", p);
	return 0;
}
