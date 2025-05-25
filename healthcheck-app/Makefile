CC = gcc
CFLAGS = -Os -Wall -Wextra -static
LDFLAGS = -static

all: healthcheck

healthcheck: healthcheck.c
	$(CC) $(CFLAGS) $(LDFLAGS) -o $@ $^

clean:
	rm -f healthcheck