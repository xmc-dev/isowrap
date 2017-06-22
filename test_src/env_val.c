#include <stdio.h>
#include <stdlib.h>

int main() {
	char *e = getenv("ISOWRAP_SPECIAL_VAL");
	if (e == NULL)
		return 1;
	else
		printf("%s", e);
	return 0;
}
