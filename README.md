# axe
A simple graph partitioning algorithm written in Go.

Designed for use for partitioning neural networks across multiple devices which has an added cost when crossing device boundaries.

Current algorithm uses a form of simulated anealing to find an approximate global minima and then once the temperature is 0 does greedy optimizination to find the local minima.

## License

Copyright (c) 2020 Tristan Rice

Licensed under the MIT license.
