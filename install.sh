#!/bin/bash
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
	PLATFORM="ubuntu"
elif [[ "$OSTYPE" == "darwin"* ]]; then
	PLATFORM="macos"
else
	echo "Unsupported OS type: $OSTYPE" >&2
	exit 1
fi

sudo wget -O /usr/local/bin/termage "https://github.com/spenserblack/termage/releases/latest/download/termage-$PLATFORM"
