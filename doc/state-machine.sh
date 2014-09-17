#!/bin/sh

STEM=state-machine

# Regenerate the figure holding the graph of states...
dot -Tpng -o ${STEM}.png ${STEM}.dot

# ...and show it
display ${STEM}.png
exit
