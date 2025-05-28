#!/bin/bash
find . -type f \( -name "*.go" -o -name "go.*" -o -name "Dockerfile" -o -name "*.sh" -o -name "README.md" \) -print0 | xargs -0 -I{} sh -c 'cat "{}"; echo "---------------------"' > project_sources.txt
echo "--------------------------" >> ./project_sources.txt
tree -d ./assets >> ./project_sources.txt

