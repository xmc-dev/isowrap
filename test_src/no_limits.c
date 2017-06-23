#include <stdio.h>
#include <stdlib.h>
#include <err.h>

int main(int argc, char **argv) {
	int e;
	if (argc < 2)
		errx(2, "not enought arguments");
	e = atoi(argv[1]);
	if (e == 0)
		printf("%s", "Success");
	else
		printf("%s", "Fail");
	return e;
}
