#!/bin/bash

for dir in $(ls -d ../*/); do
    if [ dir != '../log/' ]; then
        ln -s "$(pwd)/log.py" "${dir}log_.py"
    fi
done
