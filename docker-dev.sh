#!/bin/bash

# Development Docker Compose helper script

case "$1" in
  up|start)
    echo "Starting development environment..."
    docker compose up --build
    ;;
  down|stop)
    echo "Stopping development environment..."
    docker compose down
    ;;
  restart)
    echo "Restarting development environment..."
    docker compose down
    docker compose up --build
    ;;
  logs)
    docker compose logs -f ${2:-}
    ;;
  clean)
    echo "Cleaning up Docker resources..."
    docker compose down -v --remove-orphans
    docker system prune -f
    ;;
  *)
    echo "Usage: $0 {up|down|restart|logs [service]|clean}"
    echo ""
    echo "Commands:"
    echo "  up/start    - Start the development environment"
    echo "  down/stop   - Stop the development environment"
    echo "  restart     - Restart the development environment"
    echo "  logs        - Show logs (optionally for specific service)"
    echo "  clean       - Clean up Docker resources"
    echo ""
    echo "Examples:"
    echo "  $0 up"
    echo "  $0 logs server"
    echo "  $0 logs client"
    exit 1
    ;;
esac
