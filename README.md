# Efficient Synchronization of Linux Memory Regions over a Network: A Comparative Study and Implementation

Bachelor's thesis by Felicitas Pojtinger.

**University**: Hochschule der Medien Stuttgart<br>
**Course of Study**: Media Informatics<br>
**Date**: 2023-08-03<br>
**Academic Degree**: Bachelor of Science<br>
**Primary Supervisor**: Prof. Dr. Martin Goik<br>
**Secondary Supervisor**: M.Sc. Philip Betzler<br>

[![Deliverance CI](https://github.com/pojntfx/networked-linux-memsync/actions/workflows/deliverance.yaml/badge.svg)](https://github.com/pojntfx/networked-linux-memsync/actions/workflows/deliverance.yaml)

## Abstract

Current solutions for access, synchronization and migration of resources over a network are characterized by application-specific protocols and interfaces, which result in fragmentation and barriers to adoption. This thesis aims to address these issues by presenting a universal approach that enables direct operation on a memory region, circumventing the need for custom-built solutions. Various methods to achieve this are evaluated on parameters such as implementation overhead, initialization time, latency and throughput, and an outline of each method's architectural constraints and optimizations is provided. The proposed solution is suitable for both LAN and WAN environments, thanks to a novel approach based on block devices in user space with background push and pull mechanisms. It offers a unified API that enables mounting and migration of nearly any state over a network with minimal changes to existing applications. Illustrations of real-world use cases, configurations and backends are provided, together with a production-ready reference implementation of the full mount and migration APIs via the open-source r3map (remote mmap) library.

## Overview

This repository contains the LaTeX and Markdown markup, as well as citations, benchmarks and visualization code for the thesis. The resulting built document is **published as a PDF** by CI/CD:

<p align="center">
	<a href="https://pojntfx.github.io/networked-linux-memsync/main.pdf" rel="nofollow"><img src="./docs/thesis-badge.png" alt="Thesis badge for Pojtinger, F. (2023). Efficient Synchronization of Linux Memory Regions over a Network: A Comparative Study and Implementation" width="650"></a>
</p>

The **accompanying reference implementation** of the presented approach for working with remote memory regions and migrating them between hosts, **r3map**, can also be found on GitHub:

<p align="center">
	<a href="https://github.com/pojntfx/r3map" rel="nofollow"><img src="./docs/library-badge.png" alt="Badge for the r3map library" width="300"></a>
</p>

**Looking for an even higher-performance, production-ready library?** Check out [Loophole Labs Silo](https://github.com/loopholelabs/silo). Silo is an alternative implementation of a subset of r3map's features that focuses on performance and offers support for continuous, push-based migrations.

Additionally, you can find the thesis in different formats such as HTML and EPUB [on GitHub pages](https://pojntfx.github.io/networked-linux-memsync/), [download them from GitHub releases](https://github.com/pojntfx/networked-linux-memsync/releases/latest) or [check out the source on GitHub](https://github.com/pojntfx/networked-linux-memsync). If you want to cite this work, see [CITATION.cff](./CITATION.cff).

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

Efficient Synchronization of Linux Memory Regions over a Network: A Comparative Study and Implementation (c) 2024 Felicitas Pojtinger and contributors

SPDX-License-Identifier: Apache-2.0
