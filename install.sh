#!/bin/bash
TEMP=$(mktemp -d)
git clone https://github.com/spenserblack/termage.git $TEMP
cd $TEMP
make install
cd -
echo Removing $TEMP
rm -rf $TEMP
