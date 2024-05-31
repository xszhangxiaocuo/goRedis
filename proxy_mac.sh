#!/bin/bash

while true; do
    echo "Enter your command (or type 'exit' to quit):"
    read -r command

    if [ "$command" == "exit" ]; then
        break
    fi

    # Append CRLF to the command and send it to the server using nc
    echo -ne "$command" | nc localhost 6379
done

echo "Exiting..."
