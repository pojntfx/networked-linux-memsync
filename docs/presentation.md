---
author: [Felicitas Pojtinger]
institute: Hochschule der Medien Stuttgart
date: "2023-08-29"
subject: Efficient Synchronization of Linux Memory Regions over a Network (Presentation Notes)
keywords:
  - linux
  - memory-synchronization
  - memory-hierarchy
  - remote-memory
  - mmap
  - delta-synchronization
  - fuse
  - nbd
  - live-migration
lang: en-US
bibliography: static/references.bib
csl: static/ieee.csl
lof: true
colorlinks: false
mainfont: "Latin Modern Roman"
sansfont: "Latin Modern Roman"
monofont: "Latin Modern Mono"
code-block-font-size: \scriptsize
---

# Efficient Synchronization of Linux Memory Regions over a Network (Presentation Notes)

- Introduction
  - Title slide
  - ToC
  - About me
  - Abstract/introduction
- Methods
  - Pull-based synchronization with `userfaultfd`/Userfaults in Go with `userfaultfd`
    - Technology section: Memory organization & hierarchy
    - Technology section: Page faults
  - Push-based synchronization with `mmap` and hashing/file-based synchronization, discussion
    - Technology section: `mmap``
    - Technology section: Delta synchronization
  - Push-based synchronization with FUSE/FUSE implementation in Go, discussion
    - Technology section: FUSE
  - Mounts with NBD/NBD with go-nbd
    - Technology section: NBD
  - Push-Pull Synchronization with Mounts/managed mounts with r3map
    - Technology section: RTT, LAN and WAN
  - Pull-Based Synchronization with Migrations/Live migration
    - Technology section: Pre- and post-copy VM migration, workload analysis
- Optimizations
  - Pluggable Encryption, Authentication and Transport
  - Concurrent Backends
  - Remote Stores as Backends
  - Concurrrent RPC frameworks (dudirekta) and connection pooling (gRPC)
- Discussion and Results
  - Testing Environment
  - Access methods (userfaults vs. direct vs. managed mounts): Latency & Throughput, discussion
  - Initialization: Polling vs. udev
  - Chunking methods: Local vs. remote
  - RPC frameworks; discussion
  - Backends: Latency & throughput; discussion
  - General limitations of the r3map library (deadlocks etc.)
- Implemented Use Cases
  - Using mounts for remote swap with `ram-dl`
  - Mapping tape into memory with tapisk
- Future Use Cases
  - Improving cloud storage clients
  - Universal database, media and asset streaming
  - Universal app state mounts and migrations
- Conclusion
- Thanks
