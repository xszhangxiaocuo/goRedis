#!/bin/bash

# 检查是否提供了IP地址和端口号参数，如果没有则使用默认值
IP="${1:-localhost}"
PORT="${2:-6379}"

while true; do
    echo "Enter your command (or type 'exit' to quit):"
    read -r command

    if [ "$command" == "exit" ]; then
        break
    fi

    echo -ne "$command" | nc "$IP" "$PORT"
done

echo "Exiting..."
