#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <unistd.h>
#include <netdb.h>
#include <stdbool.h>
#include <ctype.h>
#include <sys/time.h>
#include <errno.h>

#define BUFFER_SIZE 4096
#define CONNECT_TIMEOUT_SEC 5
#define RECV_TIMEOUT_SEC 5

typedef struct {
    char host[256];
    int port;
    char endpoint[256];
    long response_time_ms;
} UrlParts;

bool parse_url(const char *url, UrlParts *parts) {
    char *host_start = strstr(url, "://");
    if (!host_start) {
        return false;
    }
    host_start += 3;
    
    char *port_start = strchr(host_start, ':');
    char *path_start = strchr(host_start, '/');

    if (port_start && (!path_start || port_start < path_start)) {
        
        size_t host_len = port_start - host_start;
        if (host_len >= sizeof(parts->host)) {
            return false;
        }
        strncpy(parts->host, host_start, host_len);
        parts->host[host_len] = '\0';

        char *port_end = path_start ? path_start : (char*)url + strlen(url);
        char port_str[16];
        size_t port_len = port_end - (port_start + 1);
        if (port_len >= sizeof(port_str)) {
            return false;
        }
        strncpy(port_str, port_start + 1, port_len);
        port_str[port_len] = '\0';
        parts->port = atoi(port_str);
    } else {        
        char *host_end = path_start ? path_start : (char*)url + strlen(url);
        size_t host_len = host_end - host_start;
        if (host_len >= sizeof(parts->host)) {
            return false;
        }
        strncpy(parts->host, host_start, host_len);
        parts->host[host_len] = '\0';
        parts->port = 80;
    }
    
    if (path_start) {
        size_t endpoint_len = strlen(path_start);
        if (endpoint_len >= sizeof(parts->endpoint)) {
            return false;
        }
        strncpy(parts->endpoint, path_start, endpoint_len);
        parts->endpoint[endpoint_len] = '\0';
    } else {
        strcpy(parts->endpoint, "/");
    }

    return true;
}

int main(int argc, char *argv[]) {
    if (argc != 2) {
        fprintf(stderr, "Usage: %s <URL> (e.g., http://example.com:8080/health)\n", argv[0]);
        return 1;
    }

    UrlParts url_parts;
    if (!parse_url(argv[1], &url_parts)) {
        fprintf(stderr, "Error: Invalid URL format. Expected: http://host[:port][/endpoint]\n");
        return 1;
    }

    printf("Host: %s\n", url_parts.host);
    printf("Port: %d\n", url_parts.port);
    printf("Endpoint: %s\n", url_parts.endpoint);

    struct addrinfo hints = {0};
    hints.ai_family = AF_INET;
    hints.ai_socktype = SOCK_STREAM;
    hints.ai_flags = AI_NUMERICSERV;

    char port_str[16];
    snprintf(port_str, sizeof(port_str), "%d", url_parts.port);

    struct addrinfo *result;
    int err = getaddrinfo(url_parts.host, port_str, &hints, &result);
    if (err != 0) {
        fprintf(stderr, "Error: Could not resolve host '%s': %s\n", url_parts.host, gai_strerror(err));
        return 1;
    }
    
    int sock = socket(result->ai_family, result->ai_socktype, result->ai_protocol);
    if (sock < 0) {
        perror("Error: Socket creation failed");
        freeaddrinfo(result);
        return 1;
    }
    
    // Set connect timeout
    struct timeval timeout;
    timeout.tv_sec = CONNECT_TIMEOUT_SEC;
    timeout.tv_usec = 0;
    if (setsockopt(sock, SOL_SOCKET, SO_SNDTIMEO, &timeout, sizeof(timeout)) < 0) {
        perror("Error: Failed to set connect timeout");
        close(sock);
        freeaddrinfo(result);
        return 1;
    }

    // Start response time measurement
    struct timeval start, end;
    gettimeofday(&start, NULL);

    if (connect(sock, result->ai_addr, result->ai_addrlen) < 0) {
        if (errno == EINPROGRESS || errno == EWOULDBLOCK) {
            fprintf(stderr, "Error: Connection timed out after %d seconds\n", CONNECT_TIMEOUT_SEC);
        } else {
            perror("Error: Connection failed");
        }
        close(sock);
        freeaddrinfo(result);
        return 1;
    }

    // Set receive timeout
    timeout.tv_sec = RECV_TIMEOUT_SEC;
    if (setsockopt(sock, SOL_SOCKET, SO_RCVTIMEO, &timeout, sizeof(timeout)) < 0) {
        perror("Error: Failed to set receive timeout");
        close(sock);
        freeaddrinfo(result);
        return 1;
    }

    freeaddrinfo(result);

    const char *request_template = "GET %s HTTP/1.1\r\n"
                                 "Host: %s\r\n"
                                 "User-Agent: healthcheck-app/1.0\r\n"
                                 "Accept: */*\r\n"
                                 "Connection: close\r\n\r\n";
    char full_request[1024];
    snprintf(full_request, sizeof(full_request), request_template, url_parts.endpoint, url_parts.host);

    if (send(sock, full_request, strlen(full_request), 0) < 0) {
        perror("Error: Failed to send request");
        close(sock);
        return 1;
    }
    
    char response[BUFFER_SIZE];
    ssize_t total_bytes = 0;
    ssize_t bytes_received;

    while ((bytes_received = recv(sock, response + total_bytes, BUFFER_SIZE - total_bytes - 1, 0)) > 0) {
        total_bytes += bytes_received;
    }

    if (bytes_received < 0) {
        perror("Error: Failed to read response");
        close(sock);
        return 1;
    }

    gettimeofday(&end, NULL);
    url_parts.response_time_ms = (end.tv_sec - start.tv_sec) * 1000 + 
                               (end.tv_usec - start.tv_usec) / 1000;

    response[total_bytes] = '\0';
    
    // Check for minimal valid HTTP response
    char *status_line = strstr(response, "HTTP/");
    if (!status_line) {
        fprintf(stderr, "Error: Invalid HTTP response\n");
        close(sock);
        return 1;
    }

    printf("Response time: %ld ms\n", url_parts.response_time_ms);
    printf("HTTP status: %.15s\n", status_line);
    
    if (!strstr(status_line, "200")) {
        fprintf(stderr, "Error: HTTP status is not 200\n");
        close(sock);
        return 1;
    }

    close(sock);
    printf("Health check passed\n");
    return 0;
}
