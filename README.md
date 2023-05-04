# Efficient Synchronization of Linux Memory Regions over a Network: A Comparative Study and Implementation

My Bachelor's Thesis: "Efficient Synchronization of Linux Memory Regions over a Network: A Comparative Study and Implementation"

[![Deliverance CI](https://github.com/pojntfx/networked-linux-memsync/actions/workflows/deliverance.yaml/badge.svg)](https://github.com/pojntfx/networked-linux-memsync/actions/workflows/deliverance.yaml)

## Overview

🚧 This project is a work-in-progress! Instructions will be added as soon as it is usable. 🚧

You can [view the notes on GitHub pages](https://pojntfx.github.io/networked-linux-memsync/), [download them from GitHub releases](https://github.com/pojntfx/networked-linux-memsync/releases/latest) or [check out the source on GitHub](https://github.com/pojntfx/networked-linux-memsync).

## Contributing

To contribute, please use the [GitHub flow](https://guides.github.com/introduction/flow/) and follow our [Code of Conduct](./CODE_OF_CONDUCT.md).

To build and open a note locally, run the following:

```shell
$ git clone https://github.com/pojntfx/networked-linux-memsync.git
$ cd networked-linux-memsync
$ ./configure
$ make depend
$ make dev-pdf/your-note # Use Bash completion to list available targets
# In another terminal
$ make open-pdf/your-note # Use Bash completion to list available targets
```

The note should now be opened. Whenever you change a source file, it will automatically be re-compiled.

## License

Efficient Synchronization of Linux Memory Regions over a Network: A Comparative Study and Implementation (c) 2023 Felicitas Pojtinger and contributors

SPDX-License-Identifier: Apache-2.0
