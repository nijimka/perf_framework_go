#!/bin/sh
set -eu

usage() {
    cat <<'EOF'
Usage: ./perf.sh [command]

  help          Show available commands (default)
  build         Build the framework locally
  run           Start the mock server on port 8080
  test          Run local performance tests using config.yaml
  clean         Remove build artifacts and reports
  docker-build  Build Docker images
  docker-run    Start the full stack in Docker
  docker-test   Run performance tests in Docker
EOF
}

if [ "$#" -gt 1 ]; then
    usage >&2
    exit 2
fi

command=${1:-help}
case "$command" in
    help|-h|--help) usage; exit 0 ;;
    build|run|test|clean|docker-build|docker-run|docker-test) ;;
    *) printf 'Unknown command: %s\n' "$command" >&2; usage >&2; exit 2 ;;
esac

project_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
cd "$project_dir"

case "$command" in
    build|run|test) dependency=go ;;
    docker-*) dependency=docker ;;
    clean) dependency= ;;
esac
if [ -n "$dependency" ] && ! command -v "$dependency" >/dev/null 2>&1; then
    printf 'Required command not found: %s\n' "$dependency" >&2
    exit 127
fi

run_docker_stack() {
    status=0
    docker compose up "$@" || status=$?
    # Copy through the Docker API so remote daemons (including DinD) work too.
    if mkdir -p reports && docker compose cp perf-test:/app/reports/. reports/; then
        echo "Reports saved to reports/"
    else
        echo "Could not copy reports from the perf-test container." >&2
        if [ "$status" -eq 0 ]; then
            status=1
        fi
    fi
    exit "$status"
}

case "$command" in
    build)
        echo "Building framework..."
        binary=bin/perf-runner
        target_os=$(go env GOOS)
        if [ "$target_os" = windows ]; then
            binary=$binary.exe
        fi
        mkdir -p bin
        go build -o "$binary" ./cmd/runner
        printf 'Build complete: %s\n' "$binary"
        ;;
    run)
        echo "Starting mock server on port 8080..."
        echo "Press Ctrl+C to stop"
        cd mock-server
        exec go run main.go
        ;;
    test)
        echo "Running performance tests..."
        exec go run ./cmd/runner -config config.yaml
        ;;
    clean)
        echo "Cleaning artifacts..."
        rm -rf bin reports
        echo "Clean complete."
        ;;
    docker-build)
        echo "Building Docker images..."
        exec docker compose build
        ;;
    docker-run)
        echo "Starting full stack in Docker..."
        run_docker_stack --build
        ;;
    docker-test)
        echo "Running tests in Docker..."
        run_docker_stack --build --force-recreate --abort-on-container-exit --exit-code-from perf-test
        ;;
esac
