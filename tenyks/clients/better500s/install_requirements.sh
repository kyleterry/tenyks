#!/bin/bash

for line in $(cat requirements.txt)
do 
    cabal install $line
done
