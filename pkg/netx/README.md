# netx

**netx** is a production-ready IPv4 and IPv6 network subnetting toolkit for Go, built exclusively on the standard library. It provides a clean, consistent API for CIDR parsing, subnet address calculation, FLSM (Fixed-Length Subnet Masking), VLSM (Variable-Length Subnet Masking), and general-purpose network utility functions.

Designed for backend infrastructure services, DevOps tooling, Kubernetes networking automation, and cloud infrastructure management.

---

## Table of Contents

- [Overview](#overview)
- [Subnetting Concepts](#subnetting-concepts)
  - [FLSM vs VLSM](#flsm-vs-vlsm)
- [Package Architecture](#package-architecture)
- [Installation](#installation)
- [API Reference](#api-reference)
  - [CIDR Parsing](#cidr-parsing)
  - [Subnet Accessors](#subnet-accessors)
  - [FLSM — Equal Subnet Splitting](#flsm--equal-subnet-splitting)
  - [VLSM — Host-Based Allocation](#vlsm--host-based-allocation)
  - [Utility Functions](#utility-functions)
- [Usage Examples](#usage-examples)
- [Real-World DevOps Scenarios](#real-world-devops-scenarios)
- [Edge Case Handling](#edge-case-handling)
- [Platform Notes](#platform-notes)
- [Engineering Guide](#engineering-guide)

---

## Overview

`netx` eliminates the boilerplate of writing subnet calculations from scratch. It addresses common pain points like:

- **CIDR parsing** — parse any CIDR and get all addressing attributes in one call
- **Subnet splitting** — divide a block into equal-sized sub-networks (FLSM)
- **Efficient allocation** — allocate right-sized subnets per host requirement (VLSM)
- **Overlap detection** — check whether two networks share any addresses
- **IPv6 support** — the same API works identically for IPv6 with \*big.Int host counts

---

## Subnetting Concepts

### CIDR Notation

A CIDR (Classless Inter-Domain Routing) address like `10.0.0.0/24` describes:

| Part                      | Meaning                                               |
| ------------------------- | ----------------------------------------------------- |
| `10.0.0.0`                | Network (base) address — all host bits are zero       |
| `/24`                     | Prefix length — number of bits in the network portion |
| `10.0.0.255`              | Broadcast address — all host bits are one             |
| `10.0.0.1` – `10.0.0.254` | Usable host range                                     |
| `254`                     | Usable host count (`2^(32-24) - 2`)                   |

### FLSM vs VLSM

|                   | FLSM                                 | VLSM                                    |
| ----------------- | ------------------------------------ | --------------------------------------- |
| **Stands for**    | Fixed-Length Subnet Masking          | Variable-Length Subnet Masking          |
| **Subnet sizes**  | All equal                            | Sized to individual requirements        |
| **Address waste** | Higher (over-allocates)              | Lower (right-sizes each subnet)         |
| **Use case**      | Simple networks, uniform departments | Complex networks with mixed host counts |
| **Example**       | Split 10.0.0.0/24 → four /26s        | Allocate /25, /26, /28 from 10.0.0.0/24 |

#### FLSM example

```
10.0.0.0/24 → split into /26:

 10.0.0.0/26    ┌────────────────────┐
                │  10.0.0.1 – .62   │  62 hosts
 10.0.0.64/26   ├────────────────────┤
                │ 10.0.0.65 – .126  │  62 hosts
 10.0.0.128/26  ├────────────────────┤
                │ 10.0.0.129 – .190 │  62 hosts
 10.0.0.192/26  └────────────────────┘
                │ 10.0.0.193 – .254 │  62 hosts
```

#### VLSM example

```
10.0.0.0/24 → DivideByHosts([100, 50, 10]):

 10.0.0.0/25   ┌────────────────────────────┐
               │ 10.0.0.1  – 10.0.0.126    │  126 hosts (satisfies 100)
 10.0.0.128/26 ├────────────────────────────┤
               │ 10.0.0.129 – 10.0.0.190   │   62 hosts (satisfies 50)
 10.0.0.192/28 ├────────────────────────────┤
               │ 10.0.0.193 – 10.0.0.206   │   14 hosts (satisfies 10)
               └────────────────────────────┘
                 10.0.0.207 – 10.0.0.255     (unused)
```

---

## Package Architecture

| File           | Responsibility                                                                     |
| -------------- | ---------------------------------------------------------------------------------- |
| `type.go`      | `Subnet` struct definition (unexported fields) and accessor methods                |
| `parse.go`     | `ParseCIDR` and `MustParseCIDR` entry points                                       |
| `subnet.go`    | Address arithmetic: network/broadcast/host-range calculation                       |
| `flsm.go`      | `Split` and `SplitIntoN` — equal-size subnet division                              |
| `vlsm.go`      | `DivideByHosts` — VLSM host-based allocation                                       |
| `utilities.go` | `Contains`, `Overlaps`, `NetworkSize`, `HostCount`, `PrefixForHosts`, `NextSubnet` |
| `doc.go`       | Package-level GoDoc documentation                                                  |

---

## Installation

```bash
go get github.com/polarixa/replify
```

Import the package:

```go
import "github.com/polarixa/replify/pkg/netx"
```

**Requirements:** Go 1.24.0 or higher. No external dependencies.

---

## API Reference

### CIDR Parsing

```go
// Parse a CIDR string and compute all addressing attributes.
sub, err := netx.ParseCIDR("192.168.1.0/24")

// Panics on invalid input — suitable for init() or tests.
sub := netx.MustParseCIDR("10.0.0.0/8")
```

### Subnet Accessors

All `Subnet` fields are unexported. Read them through accessor methods:

```go
sub, _ := netx.ParseCIDR("192.168.1.0/24")

sub.IPNet()            *net.IPNet     // underlying net.IPNet
sub.NetworkAddress()   net.IP         // 192.168.1.0
sub.BroadcastAddress() net.IP         // 192.168.1.255
sub.FirstHost()        net.IP         // 192.168.1.1
sub.LastHost()         net.IP         // 192.168.1.254
sub.TotalHosts()       *big.Int       // 254
sub.Prefix()           int            // 24
sub.String()           string         // "192.168.1.0/24"
```

### FLSM — Equal Subnet Splitting

```go
base := netx.MustParseCIDR("10.0.0.0/24").IPNet()

// Split into equal /26 subnets
subnets, err := netx.Split(base, 26)
// → [10.0.0.0/26, 10.0.0.64/26, 10.0.0.128/26, 10.0.0.192/26]

// Split into exactly N equal parts (N must be a power of 2)
subnets, err = netx.SplitIntoN(base, 4)
// → same four /26 subnets

// Convert to strings
strs := netx.SubnetsToStrings(subnets)
```

### VLSM — Host-Based Allocation

```go
base := netx.MustParseCIDR("10.0.0.0/24").IPNet()

// Allocate subnets sized to satisfy each host requirement.
// Requirements are automatically sorted largest-first.
subnets, err := netx.DivideByHosts(base, []int{100, 50, 10})
// subnets[0]: 10.0.0.0/25   — 126 usable hosts (satisfies 100)
// subnets[1]: 10.0.0.128/26 —  62 usable hosts (satisfies  50)
// subnets[2]: 10.0.0.192/28 —  14 usable hosts (satisfies  10)

// Convert to strings
strs := netx.AllocatedSubnetsToStrings(subnets)
```

### Utility Functions

```go
// Check whether an IP belongs to a network
_, n, _ := net.ParseCIDR("10.0.0.0/8")
netx.Contains(n, net.ParseIP("10.1.2.3"))  // true

// Check whether two networks overlap
subA := netx.MustParseCIDR("10.0.0.0/24").IPNet()
subB := netx.MustParseCIDR("10.0.0.128/25").IPNet()
netx.Overlaps(subA, subB)  // true

// Total address count (including network + broadcast)
netx.NetworkSize(n)  // *big.Int — e.g. 256 for /24

// Usable host count for a prefix
netx.HostCount(24, 32)  // 254 (IPv4 /24)
netx.HostCount(31, 32)  // 2   (RFC 3021 point-to-point)
netx.HostCount(32, 32)  // 1   (single host)

// Smallest prefix providing ≥ N usable hosts
netx.PrefixForHosts(100, 32)  // 25 → /25 provides 126 hosts
netx.PrefixForHosts(254, 32)  // 24 → /24 provides 254 hosts

// Next contiguous subnet of a given prefix
base := netx.MustParseCIDR("10.0.0.0/26").IPNet()
next, _ := netx.NextSubnet(base, 26)
fmt.Println(next)  // "10.0.0.64/26"
```

---

## Usage Examples

### Parse a CIDR and print its attributes

```go
package main

import (
    "fmt"
    "github.com/polarixa/replify/pkg/netx"
)

func main() {
    sub, err := netx.ParseCIDR("10.128.0.0/18")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Network:    %s\n", sub.NetworkAddress())
    fmt.Printf("Broadcast:  %s\n", sub.BroadcastAddress())
    fmt.Printf("First host: %s\n", sub.FirstHost())
    fmt.Printf("Last host:  %s\n", sub.LastHost())
    fmt.Printf("Hosts:      %s\n", sub.TotalHosts())
    fmt.Printf("Prefix:     /%d\n", sub.Prefix())
}
// Network:    10.128.0.0
// Broadcast:  10.191.255.255
// First host: 10.128.0.1
// Last host:  10.191.255.254
// Hosts:      16382
// Prefix:     /18
```

### Split a block into equal subnets (FLSM)

```go
base := netx.MustParseCIDR("172.16.0.0/20").IPNet()
subs, err := netx.Split(base, 24)
if err != nil {
    panic(err)
}
for _, s := range subs {
    fmt.Println(s)
}
// 172.16.0.0/24
// 172.16.1.0/24
// ...
// 172.16.15.0/24
```

### Allocate right-sized subnets per department (VLSM)

```go
base := netx.MustParseCIDR("10.0.0.0/22").IPNet()
subnets, err := netx.DivideByHosts(base, []int{500, 200, 100, 50, 25})
if err != nil {
    panic(err)
}
for _, s := range subnets {
    fmt.Printf("%s  hosts=%d\n", s.String(), s.TotalHosts())
}
```

### Detect overlapping networks

```go
networks := []string{
    "10.0.0.0/24",
    "10.0.0.128/25",
    "192.168.1.0/24",
}
parsed := make([]*net.IPNet, len(networks))
for i, c := range networks {
    parsed[i] = netx.MustParseCIDR(c).IPNet()
}
for i := 0; i < len(parsed); i++ {
    for j := i + 1; j < len(parsed); j++ {
        if netx.Overlaps(parsed[i], parsed[j]) {
            fmt.Printf("OVERLAP: %s ↔ %s\n", networks[i], networks[j])
        }
    }
}
// OVERLAP: 10.0.0.0/24 ↔ 10.0.0.128/25
```

---

## Real-World DevOps Scenarios

### IP Address Planning

```go
// Divide a company's 10.0.0.0/16 block across regions and teams.
corporate := netx.MustParseCIDR("10.0.0.0/16").IPNet()
regions, _ := netx.SplitIntoN(corporate, 4)  // four /18s

for i, region := range regions {
    teams, _ := netx.Split(region, 24)  // each /18 → 64 x /24
    fmt.Printf("Region %d: %d team networks allocated\n", i+1, len(teams))
}
```

### Kubernetes Cluster Networking

```go
// Allocate pod and service CIDRs from a cluster block.
clusterCIDR := netx.MustParseCIDR("100.64.0.0/14").IPNet()

// Each node gets a /24 pod CIDR; service range is a /20.
subnets, err := netx.DivideByHosts(clusterCIDR, []int{
    65534, // node pod pool (a /16)
    4094,  // service CIDR  (a /20)
})
if err != nil {
    log.Fatal(err)
}
fmt.Println("Pod pool:     ", subnets[0].String())
fmt.Println("Service CIDR: ", subnets[1].String())
```

### Cloud Infrastructure Automation

```go
// Validate that a user-supplied CIDR is within the VPC's address space.
vpc := netx.MustParseCIDR("172.31.0.0/16").IPNet()
userInput := "172.31.5.0/24"

sub, err := netx.ParseCIDR(userInput)
if err != nil {
    log.Fatalf("invalid CIDR: %v", err)
}
if !netx.Contains(vpc, sub.NetworkAddress()) {
    log.Fatalf("subnet %s is outside the VPC %s", userInput, vpc)
}
fmt.Println("Subnet is within the VPC ✓")
```

### Network Monitoring and Inventory

```go
// Report all hosts in a monitored range.
sub := netx.MustParseCIDR("192.168.10.0/27")
fmt.Printf("Monitoring range: %s – %s (%d hosts)\n",
    sub.FirstHost(), sub.LastHost(), sub.TotalHosts())
```

---

## Edge Case Handling

| Scenario                           | Behaviour                                                       |
| ---------------------------------- | --------------------------------------------------------------- |
| `/31` network (RFC 3021)           | `TotalHosts()` = 2; `FirstHost()` = network address             |
| `/32` single host                  | `TotalHosts()` = 1; `FirstHost()` = `LastHost()` = host address |
| IPv6 subnet                        | All accessors work; `TotalHosts()` returns `*big.Int`           |
| Insufficient space (VLSM)          | `DivideByHosts` returns a descriptive error                     |
| `nil` arguments                    | All utility functions return safe zero values or errors         |
| Invalid CIDR                       | `ParseCIDR` returns a wrapped error; `MustParseCIDR` panics     |
| Non-power-of-2 `n` in `SplitIntoN` | Returns an error                                                |

---

## Platform Notes

`netx` depends only on `net`, `math/big`, and `sort` from the Go standard library and produces identical results on Linux, macOS, and Windows.

No OS-specific code paths exist in this package.

## Engineering Guide

For a comprehensive technical reference covering real-world subnetting
scenarios, binary-level calculations, efficiency analysis, VLSM vs. FLSM
comparison, route summarization, multi-VLAN design, and algorithm pseudocode
for all major subnetting operations, see:

**[SUBNETTING_GUIDE.md](./enterprise-grades/SUBNETTING_GUIDE.md)**

Topics covered in the guide:

| Section              | Content                                                                   |
| -------------------- | ------------------------------------------------------------------------- |
| Background           | Classful addressing → CIDR transition                                     |
| Bitwise logic        | Subnet mask AND/OR operations, binary examples                            |
| Power-of-two rule    | Standard formula, /31 and /32 edge cases                                  |
| FLSM walkthrough     | `192.168.10.0/24` divided into 4 equal /26 subnets                        |
| VLSM walkthrough     | `10.0.0.0/24` allocated for Sales/IT/HR/P2P with 73% less waste than FLSM |
| IPv4 conservation    | `/29` public block assignment and NAT strategies                          |
| Route summarization  | Binary LCP, aggregating four /24s into a /22                              |
| Multi-VLAN design    | VLAN-to-subnet alignment, gateway assignment, security policy             |
| Algorithm pseudocode | FLSM split, VLSM allocation, route summarization, prefix search           |
| API mapping          | Every concept mapped to the corresponding `netx` Go function              |

---
