# Robohash for Golang

[![Go Report Card](https://goreportcard.com/badge/github.com/terem42/robohash)](https://goreportcard.com/report/github.com/terem42/robohash)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Golang implementation of Robohash, the awesome library for generating unique robot/avatar images from any text hash. This is a port of the original [Robohash](https://github.com/e1ven/Robohash) project with performance improvements and additional features, such as performance improvements, image caching and AVIF/WebP support

Alllows image returned being encoded either PNG, lossless WEBP or AVIF. PNG format is used by default, AVIF or WEBP when either .avif or .webp extensions are supplied

Available as a module or standalone HTTP server.

**Note**: This project uses the original image assets from Robohash under their [original license](https://github.com/e1ven/Robohash/blob/master/LICENSE).

## Features

- 🚀 High-performance image generation (5-10x faster than original Python version)
- 🖼️ Supports all 5 original sets: Robots, Monsters, Heads, Cats, and Human Avatars
- 🎨 Customizable size, background, and set selection
- ⚡ Built with Go's standard libraries (no external image processing dependencies)
- 🐳 Docker-ready for easy deployment

## Installation

```bash
go get github.com/terem42/robohash
```

Or using Docker for standalone HTTP server version:

```bash
docker pull ghcr.io/terem42/robohash
docker run -p 8080:8080 ghcr.io/terem42/robohash
```

## Usage

### Basic URL Format

```
http://yourserver.com/{TEXT}.png?{PARAMETERS}
http://yourserver.com/{TEXT}.avif?{PARAMETERS}
```

### Examples

1. **Simple robot avatar**:
   ```
   https://robohash.yourserver.com/alice.png
   ```

2. **Monster avatar with custom size**:
   ```
   https://robohash.yourserver.com/bob.png?set=set2&size=200x200
   ```

3. **Robot head with blue background**:
   ```
   https://robohash.yourserver.com/charlie.png?set=set3&bgset=bg1
   ```

4. **Human avatar**:
   ```
   https://robohash.yourserver.com/dave@email.com.png?set=set5
   ```

5. **Human avatar encoded as AVIF**:
   ```
   https://robohash.yourserver.com/dave@email.com.avif?set=set5
   ```   
6. **Human avatar encoded as WEBP**:
   ```
   https://robohash.yourserver.com/dave@email.com.webp?set=set5
   ```   


### Available Parameters

| Parameter | Values | Description |
|-----------|--------|-------------|
| `set`     | set1, set2, set3, set4, set5 | Image set to use (default: set1) |
| `size`    | {width}x{height} | Output dimensions (e.g., 300x300) |
| `bgset`   | bg1, bg2 | Background set (only for sets 1-3) |

## Sets Overview

1. **Set1 (Robots)** - 300×300px  
   Colorful robot avatars with multiple parts
   ```
   /text.png?set=set1
   ```

2. **Set2 (Monsters)** - 350×350px  
   Scary monster illustrations
   ```
   /text.png?set=set2&bgset=bg1
   ```

3. **Set3 (Heads)** - 1015×1015px  
   Detailed robot heads (white background recommended)
   ```
   /text.png?set=set3&size=500x500
   ```

4. **Set4 (Cats)** - 1024×1024px  
   Adorable cat avatars
   ```
   /text.png?set=set4
   ```

5. **Set5 (Humans)** - 1024×1024px  
   Diverse human avatars with clothing options
   ```
   /text.png?set=set5
   ```

## API Integration

```go
package main

import (
	"github.com/terem42/robohash/robohash"
)

func main() {
	// Create a new Robohash instance
	rh := robohash.NewRoboHash("alice", robohash.Set3)
	
	// Generate image
	img, err := rh.Generate()
	if err != nil {
		panic(err)
	}
	
   // rest of the code
}
```

## HTTP Caching Headers

The server automatically adds optimal caching headers for generated images  
Content-Type is set, depending on returned image  
Example for PNG images

```http
HTTP/1.1 200 OK
Cache-Control: public, max-age=31536000
ETag: "a1b2c3d4e5f6..."
Last-Modified: Wed, 21 Oct 2023 07:28:00 GMT
Content-Type: image/png
Content-Length: 24872
```

## Decoded PNG assets cache

to significantly speed up image generation, package uses internal PNG assets image memory caching, both original and resized

  - Stores parsed source images in memory
  - LRU eviction policy
  - Key format: `path|widthxheight` (e.g. `assets/set1/blue/003#01Body/5.png|300x300`)

The cache size can be configured using environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ROBOHASH_IMG_CACHE_SIZE` | 100 | Maximum image cache size in megabytes |

Example:
```bash
# Set cache size to 100MB
export ROBOHASH_IMG_CACHE_SIZE=100
docker run -e ROBOHASH_IMG_CACHE_SIZE=100 -p 8080:8080 ghcr.io/terem42/robohash


## Nginx Configuration

Example production configuration with caching:

```nginx
proxy_cache_path /var/cache/nginx/robohash 
    levels=1:2 
    keys_zone=robohash_cache:10m
    inactive=1y
    max_size=1g;

server {
    location / {
        proxy_pass http://localhost:8080;
        proxy_cache robohash_cache;
        proxy_cache_valid 200 1y;
        proxy_cache_use_stale error timeout updating;
        add_header X-Cache-Status $upstream_cache_status;
    }
}
```

## Deployment

1. **Standalone binary with embedded resources**:
   ```bash
   go build -o robohash-go ./cmd/server
   ./robohash-go
   ```

2. **Docker**:
   ```bash
   docker build -t ghcr.io/terem42/robohash .
   docker run -p 8080:8080 ghcr.io/terem42/robohash
   ```

3. **Kubernetes**:
   ```yaml
   # Sample deployment.yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: robohash
   spec:
     replicas: 3
     template:
       spec:
         containers:
         - name: robohash
           image: ghcr.io/terem42/robohash
           ports:
           - containerPort: 8080
   ```

## Credits

This project uses the original image assets from [Robohash](https://github.com/e1ven/Robohash) by Colin Davis (e1ven), available under the MIT License.

## License

MIT © Andrey Prokopenko
```
