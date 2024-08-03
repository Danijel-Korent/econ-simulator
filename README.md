# Refactor

## File splitting

The code has been split into 5 additional files (besides main.go). 

### people.go

This file now handles everything to do with people. It contains the struct definitions and related functions.

### producers.go

This file is responsible for everything to do with producers, much like people.go

### output.go

This one remains unchanged. It handles the terminal and HTML output

### config.go

This file handles the loading of the configuration, the creation of the default config, and everything related to that. The entire process of loading/creating the configuration has been split into multiple functions for clarity.

### util.go

This file is the smallest. It simply contains small maths functions (e.g. random number generation, distance calculation). While right now it seems pointless, it may become useful to split out if the simulation is going to expand.

## Major changes

The actions of people and producers each simulation step are now handled in their respective files in functions called `simulationStep`. The main `simulationStep` function in `main.go` now only serves to call these functions on each producer and person.
