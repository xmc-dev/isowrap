#include <err.h>
#include <stdio.h>
#include <signal.h>
#include <stdlib.h>
#include <unistd.h>

#include <sys/wait.h>

int main(int argc, char **argv) {
	int i, max_proc;
	/* prevent zombies */
	signal(SIGCHLD, SIG_IGN);

	if (argc < 2)
		errx(1, "not enough arguments");
	max_proc = atoi(argv[1]);

	for (i = 0; i < max_proc; i++) {
		pid_t p = fork();
		if (p == -1) {
			err(1, "fork()");
		} else if (p == 0) {
			printf("%d ", i);
			exit(0);
		}
	}
	wait(NULL);

	return 0;
}
