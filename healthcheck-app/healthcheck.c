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

#define BUFFER_SIZE 4096

typedef struct {
    char host[256];
    int port;
    char endpoint[256];
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
    
    if (connect(sock, result->ai_addr, result->ai_addrlen) < 0) {
        perror("Error: Connection failed");
        close(sock);
        freeaddrinfo(result);
        return 1;
    }

    freeaddrinfo(result);

    const char *request_template = "GET %s HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n";
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

    response[total_bytes] = '\0';
    printf("Server response:\n%s\n", response);
    
    if (!strstr(response, "200 OK")) {
        fprintf(stderr, "Error: HTTP status is not 200 OK\n");
        close(sock);
        return 1;
    }

    close(sock);
    return 0;
}