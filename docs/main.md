---
author: [Felicitas Pojtinger (Stuttgart Media University)]
date: "2023-08-04"
subject: "Efficient Synchronization of Linux Memory Regions over a Network: A Comparative Study and Implementation"
keywords:
  - linux memory management
  - userfaultfd
  - mmap
  - inotify
  - hash-based change detection
  - delta synchronization
  - msync
  - custom filesystem
  - nbd protocol
  - performance evaluation
subtitle: "A user-friendly approach to application-agnostic state synchronization"
lang: en-US
abstract: |
  ## \abstractname{} {.unnumbered .unlisted}

  This study presents a comprehensive comparison and implementation of various methods for synchronizing memory regions in Linux systems over a network. Four approaches are evaluated: (1) handling page faults in userspace with `userfaultfd`, (2) utilizing `mmap` for change notifications, (3) hash-based change detection, and (4) custom filesystem implementation. Each option is thoroughly examined in terms of implementation, performance, and associated trade-offs. The study culminates in a summary that compares the options based on ease of implementation, CPU load, and network traffic, and offers recommendations for the optimal solution depending on the specific use case, such as data change frequency and kernel/OS compatibility.
bibliography: static/references.bib
csl: static/ieee.csl
---

# Efficient Synchronization of Linux Memory Regions over a Network: A Comparative Study and Implementation

🚧 This project is a work-in-progress! Instructions will be added as soon as it is usable. 🚧
