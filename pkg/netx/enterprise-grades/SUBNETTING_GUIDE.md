# Real-World Subnetting: A Technical Engineering Guide

A professional reference manual for IPv4 and IPv6 subnet design, covering
production networking scenarios, binary-level calculations, efficiency
analysis, and algorithmic logic suitable for implementation in a Golang
networking library such as `netx`.

---

## Table of Contents

1. [Background: Classful vs. Classless Addressing](#1-background-classful-vs-classless-addressing)
2. [Core Concepts: Subnet Mask Bitwise Logic](#2-core-concepts-subnet-mask-bitwise-logic)
3. [The Power-of-Two Rule and Edge Cases](#3-the-power-of-two-rule-and-edge-cases)
4. [Scenario 1 — FLSM: Static Equal Subnetting](#4-scenario-1--flsm-static-equal-subnetting)
5. [Scenario 2 — VLSM: Variable Length Subnet Masking](#5-scenario-2--vlsm-variable-length-subnet-masking)
6. [Scenario 3 — Public IPv4 Conservation](#6-scenario-3--public-ipv4-conservation)
7. [Scenario 4 — Route Summarization (Supernetting)](#7-scenario-4--route-summarization-supernetting)
8. [Scenario 5 — Multi-VLAN Subnet Design](#8-scenario-5--multi-vlan-subnet-design)
9. [Algorithm Pseudocode Reference](#9-algorithm-pseudocode-reference)
10. [Mapping Concepts to the `netx` Go API](#10-mapping-concepts-to-the-netx-go-api)

---

## 1. Background: Classful vs. Classless Addressing

### Historical Classful Addressing

Before 1993, IPv4 addresses were divided into fixed classes based on the
leading bits of the first octet.

| Class | First Octet Range | Default Mask | Networks  | Hosts/Net  |
| ----- | ----------------- | ------------ | --------- | ---------- |
| A     | 1–126             | /8           | 126       | 16,777,214 |
| B     | 128–191           | /16          | 16,384    | 65,534     |
| C     | 192–223           | /24          | 2,097,152 | 254        |

**The problem:** a company needing 300 hosts received a full Class B (65,534
hosts), wasting over 65,000 addresses. There was no middle ground. The global
routing table grew rapidly because each organisation required a separate
routing entry.

### CIDR: Classless Inter-Domain Routing (RFC 1519, 1993)

CIDR replaced the class system with a variable-length prefix notation:
`A.B.C.D/prefix-length`.

The prefix length can be any value from /0 to /32 (IPv4) or /128 (IPv6),
completely decoupled from octet boundaries.

**Benefits of CIDR:**

- **Address efficiency**: allocate exactly the block size an organisation
  needs (e.g. /27 for 30 hosts).
- **Route aggregation**: an ISP can advertise a single `/20` covering 16
  internal `/24` blocks, reducing global routing table size.
- **Hierarchical delegation**: large blocks are divided top-down, enabling
  efficient summarisation at every tier.

**Binary representation of CIDR:**

```
Address:  192.168.1.0/24
Binary:   11000000.10101000.00000001.00000000
                                              ^^^^^^^^ ← 8 host bits
          ^^^^^^^^^^^^^^^^^^^^^^^^ ← 24 network bits
```

---

## 2. Core Concepts: Subnet Mask Bitwise Logic

### What Is a Subnet Mask?

A subnet mask is a 32-bit (IPv4) or 128-bit (IPv6) value where the leading
`n` bits are all 1 (the network portion) and the remaining bits are all 0
(the host portion).

A `/24` mask in binary:

```
11111111.11111111.11111111.00000000
= 255.255.255.0
```

A `/26` mask in binary:

```
11111111.11111111.11111111.11000000
= 255.255.255.192
```

### Deriving the Network Address (Bitwise AND)

The network address is computed by performing a bitwise AND between the IP
address and the subnet mask. This zeroes out all host bits.

**Example:**

```
IP Address:    192.168.1.130
               11000000.10101000.00000001.10000010
Subnet Mask:   255.255.255.0
               11111111.11111111.11111111.00000000
                                         AND
Result:        192.168.1.0
               11000000.10101000.00000001.00000000
```

Every device on the same `/24` network produces the same result when its
address is ANDed with the mask, confirming they share the same network
segment.

### Deriving the Broadcast Address (Bitwise OR with Inverted Mask)

The broadcast address is computed by ORing the network address with the
bitwise complement (NOT) of the mask — setting all host bits to 1.

**Example (`192.168.1.0/24`):**

```
Network:        11000000.10101000.00000001.00000000
Inverted mask:  00000000.00000000.00000000.11111111
                                           OR
Broadcast:      11000000.10101000.00000001.11111111
              = 192.168.1.255
```

### Go `netx` internal implementation

```go
func computeBroadcast(network net.IP, mask net.IPMask) net.IP {
    broadcast := make(net.IP, len(network))
    for i := range network {
        broadcast[i] = network[i] | ^mask[i]  // OR with inverted mask
    }
    return broadcast
}
```

---

## 3. The Power-of-Two Rule and Edge Cases

### Standard Formula

For a subnet with prefix `/p` in a `b`-bit address family:

```
Total addresses  = 2^(b - p)
Usable hosts     = 2^(b - p) - 2
```

The two reserved addresses are:

- **Network address** (all host bits = 0) — identifies the subnet itself.
- **Broadcast address** (all host bits = 1) — frames sent here are received
  by every device on the subnet.

**IPv4 Examples:**

| Prefix | Host Bits | Total Addresses | Usable Hosts |
| ------ | --------- | --------------- | ------------ |
| /24    | 8         | 256             | 254          |
| /25    | 7         | 128             | 126          |
| /26    | 6         | 64              | 62           |
| /27    | 5         | 32              | 30           |
| /28    | 4         | 16              | 14           |
| /29    | 3         | 8               | 6            |
| /30    | 2         | 4               | 2            |

### Edge Case: /31 — Point-to-Point Links (RFC 3021)

A `/31` has only 2 total addresses. Applying the standard formula yields
`2 - 2 = 0` usable hosts, which would be impractical. RFC 3021 defines /31
for point-to-point links where both addresses are usable host addresses —
there is no separate network or broadcast address.

```
10.0.0.0/31
  10.0.0.0  → Host A (router interface)
  10.0.0.1  → Host B (router interface)
```

Usable hosts = **2** (exception to the standard formula).

### Edge Case: /32 — Host Route

A `/32` identifies exactly one address — the host itself. Used for loopback
interfaces and policy routing.

```
10.0.0.1/32
  First host = Last host = 10.0.0.1
  Total usable hosts = 1
```

### `netx` implementation of edge cases

```go
func computeTotalHosts(prefix, bits int) *big.Int {
    hostBits := bits - prefix
    switch {
    case hostBits == 0:   // /32 or /128
        return big.NewInt(1)
    case hostBits == 1:   // /31 or /127
        return big.NewInt(2)
    default:
        total := new(big.Int).Lsh(big.NewInt(1), uint(hostBits)) // 2^hostBits
        total.Sub(total, big.NewInt(2))                           // minus network + broadcast
        return total
    }
}
```

---

## 4. Scenario 1 — FLSM: Static Equal Subnetting

### Problem Description

A company receives the block `192.168.10.0/24` (256 addresses). The network
must be divided into **4 equal subnets** for branches in London, Paris,
Tokyo, and Sydney.

**Input parameters:**

- Base network: `192.168.10.0/24`
- Number of subnets: 4
- Strategy: FLSM (all subnets identical size)

### Step 1: Determine How Many Bits to Borrow

To create `N` equal subnets, borrow the smallest number of bits `n` such
that `2^n >= N`.

```
N = 4  →  2^n >= 4  →  n = 2
```

New prefix = original prefix + borrowed bits = 24 + 2 = **/26**

### Step 2: Calculate New Subnet Parameters

```
Host bits remaining = 32 - 26 = 6
Addresses per subnet = 2^6 = 64
Usable hosts per subnet = 64 - 2 = 62
```

### Step 3: Enumerate All Subnets

The subnets advance in increments of `2^host_bits = 64`:

| Subnet     | CIDR              | Network Addr   | First Host     | Last Host      | Broadcast      | Usable Hosts |
| ---------- | ----------------- | -------------- | -------------- | -------------- | -------------- | ------------ |
| 1 (London) | 192.168.10.0/26   | 192.168.10.0   | 192.168.10.1   | 192.168.10.62  | 192.168.10.63  | 62           |
| 2 (Paris)  | 192.168.10.64/26  | 192.168.10.64  | 192.168.10.65  | 192.168.10.126 | 192.168.10.127 | 62           |
| 3 (Tokyo)  | 192.168.10.128/26 | 192.168.10.128 | 192.168.10.129 | 192.168.10.190 | 192.168.10.191 | 62           |
| 4 (Sydney) | 192.168.10.192/26 | 192.168.10.192 | 192.168.10.193 | 192.168.10.254 | 192.168.10.255 | 62           |

### Step 4: Binary Verification

Verify subnet 2 (`192.168.10.64/26`):

```
Network:   11000000.10101000.00001010.01000000  (192.168.10.64)
Mask:      11111111.11111111.11111111.11000000  (255.255.255.192)
Broadcast: 11000000.10101000.00001010.01111111  (192.168.10.127)
                                       ^^^^^^ ← 6 host bits
```

### Efficiency Analysis

```
Total addresses in /24:        256
Addresses used (4 × 64):       256
Wasted addresses:               0
Utilisation:                  100%
```

FLSM achieves 100% address efficiency **only when** the number of subnets is
an exact power of 2 AND all subnets have identical host requirements. If any
branch needs fewer than 62 hosts, the unused addresses within each /26 are
wasted.

---

## 5. Scenario 2 — VLSM: Variable Length Subnet Masking

### Problem Description

A network engineer must allocate subnets from `10.0.0.0/24` for the
following departments:

| Department | Hosts Required |
| ---------- | -------------- |
| Sales      | 100            |
| IT         | 20             |
| HR         | 5              |
| P2P Link A | 2              |
| P2P Link B | 2              |

FLSM would require a /25 for each department (126 usable hosts). Five /25
blocks would require a /22, but the available block is only /24 —
insufficient and extremely wasteful. VLSM solves this.

### Step 1: Sort Requirements Largest First

Allocating largest subnets first guarantees sequential non-overlapping
alignment.

```
Sorted requirements: 100, 20, 5, 2, 2
```

### Step 2: Find Minimum Prefix for Each Requirement

For each requirement `h`, find the smallest prefix `p` (largest block) such
that `2^(32-p) - 2 >= h`.

```
h = 100  →  2^7 - 2 = 126 >= 100  →  /25  (128 addresses)
h = 20   →  2^5 - 2 =  30 >= 20   →  /27  (32 addresses)
h = 5    →  2^3 - 2 =   6 >= 5    →  /29  (8 addresses)
h = 2    →  2^2 - 2 =   2 >= 2    →  /30  (4 addresses)
h = 2    →  2^2 - 2 =   2 >= 2    →  /30  (4 addresses)
```

### Step 3: Allocate Subnets Sequentially

Starting at `10.0.0.0`, place each subnet immediately after the previous:

| Dept      | CIDR          | Network Addr | First Host | Last Host  | Broadcast  | Usable | Alloc |
| --------- | ------------- | ------------ | ---------- | ---------- | ---------- | ------ | ----- |
| Sales /25 | 10.0.0.0/25   | 10.0.0.0     | 10.0.0.1   | 10.0.0.126 | 10.0.0.127 | 126    | 128   |
| IT /27    | 10.0.0.128/27 | 10.0.0.128   | 10.0.0.129 | 10.0.0.158 | 10.0.0.159 | 30     | 32    |
| HR /29    | 10.0.0.160/29 | 10.0.0.160   | 10.0.0.161 | 10.0.0.166 | 10.0.0.167 | 6      | 8     |
| P2P A /30 | 10.0.0.168/30 | 10.0.0.168   | 10.0.0.169 | 10.0.0.170 | 10.0.0.171 | 2      | 4     |
| P2P B /30 | 10.0.0.172/30 | 10.0.0.172   | 10.0.0.173 | 10.0.0.174 | 10.0.0.175 | 2      | 4     |

Addresses consumed: 128 + 32 + 8 + 4 + 4 = **176**  
Addresses remaining in /24: 256 − 176 = **80** (unallocated, available for future growth)

### Step 4: Efficiency Comparison

| Strategy | Addresses Needed | Addresses Required | Waste     |
| -------- | ---------------- | ------------------ | --------- |
| FLSM /25 | 129              | 5 × 128 = 640      | 511 (79%) |
| VLSM     | 129              | 176                | 47 (27%)  |

**VLSM uses 176 addresses vs. FLSM's 640 — a 73% reduction in allocation.**
The remaining 80 addresses in the /24 can support future subnets without
requiring a new block.

### Why VLSM Works

VLSM exploits the hierarchical nature of binary addressing. A /25 "uses up"
the first half of the /24; within the second half (the /25 starting at
.128), a /27, /29, and two /30s perfectly pack into the first 48 addresses
of that half, leaving the rest free. No address space is wasted between
subnets because sequential allocation with largest-first ordering ensures
natural alignment.

---

## 6. Scenario 3 — Public IPv4 Conservation

### Problem Description

An ISP allocates the block `203.0.113.8/29` to a small business for their
internet-facing infrastructure.

### Address Space Analysis

```
Prefix:          /29
Host bits:       32 - 29 = 3
Total addresses: 2^3 = 8
Usable hosts:    2^3 - 2 = 6

Network address:   203.0.113.8
Broadcast address: 203.0.113.15
First usable:      203.0.113.9
Last usable:       203.0.113.14
```

Binary layout:

```
203.0.113.8   = 11001011.00000000.01110001.00001000
203.0.113.15  = 11001011.00000000.01110001.00001111
                                                 ^^^  ← 3 host bits
```

### Typical Address Assignments

| Address      | Role                               |
| ------------ | ---------------------------------- |
| 203.0.113.8  | Network address (reserved)         |
| 203.0.113.9  | Upstream router (ISP gateway)      |
| 203.0.113.10 | Edge firewall WAN interface        |
| 203.0.113.11 | Load balancer VIP                  |
| 203.0.113.12 | Spare / secondary public interface |
| 203.0.113.13 | Spare / secondary public interface |
| 203.0.113.14 | Spare / monitoring or secondary LB |
| 203.0.113.15 | Broadcast address (reserved)       |

### Conservation Strategies

**NAT (Network Address Translation):** The firewall at `203.0.113.10`
translates all internal RFC 1918 addresses (e.g. `10.0.0.0/8`) to a single
public IP, multiplying the usable capacity from 6 public IPs to thousands of
private hosts.

**IP Reservation Planning:**

- Reserve one address for the upstream ISP gateway (required for BGP/routing
  sessions).
- Reserve one or two addresses for failover or secondary interfaces.
- Avoid assigning all 6 without headroom — address changes require ISP
  coordination.

**Load balancing:** Assign a single public VIP to the load balancer; the
private server farm uses RFC 1918 space behind NAT. This makes the entire
server pool reachable through one public address.

---

## 7. Scenario 4 — Route Summarization (Supernetting)

### Problem Description

A router in an autonomous system knows about four contiguous networks:

```
192.168.4.0/24
192.168.5.0/24
192.168.6.0/24
192.168.7.0/24
```

Rather than advertising four separate routes to the upstream ISP, the
network engineer wants to advertise a single summary route.

### Step 1: Write All Network Addresses in Binary

```
192.168.4.0  = 11000000.10101000.00000100.00000000
192.168.5.0  = 11000000.10101000.00000101.00000000
192.168.6.0  = 11000000.10101000.00000110.00000000
192.168.7.0  = 11000000.10101000.00000111.00000000
```

### Step 2: Find the Longest Common Prefix

Align the binary representations and find where they first diverge:

```
              ← 22 bits match →  diverge
11000000.10101000.000001  00 .00000000
11000000.10101000.000001  01 .00000000
11000000.10101000.000001  10 .00000000
11000000.10101000.000001  11 .00000000
```

The first 22 bits are identical. The 23rd and 24th bits cover all four
combinations (00, 01, 10, 11), confirming all four /24 networks fit within
a single /22.

### Step 3: Construct the Summary Route

```
Common prefix: 11000000.10101000.000001xx  (22 bits)
Summary:       192.168.4.0/22
```

**Verification:**

```
Network:   192.168.4.0
Mask:      255.255.252.0   (/22)
Range:     192.168.4.0 – 192.168.7.255
Covers:    192.168.4.0/24, 192.168.5.0/24, 192.168.6.0/24, 192.168.7.0/24 ✓
```

### Benefits of Route Summarization

| Benefit                  | Explanation                                                      |
| ------------------------ | ---------------------------------------------------------------- |
| Smaller routing tables   | One entry replaces four; scales from 4 routes to any power of 2  |
| Faster routing decisions | Fewer lookups; hardware TCAM entries are preserved               |
| Stability isolation      | Flapping of an internal /24 does not propagate to the backbone   |
| Scalability              | ISPs aggregate thousands of customer routes into a single prefix |

**Prerequisite for summarisation:** the subnets to be aggregated must be
**contiguous** and **start on a boundary** aligned to the summary prefix.
`192.168.5.0/24 – 192.168.7.0/24` (3 subnets) cannot form a clean summary
because 3 is not a power of 2.

---

## 8. Scenario 5 — Multi-VLAN Subnet Design

### Problem Description

An enterprise network receives the block `10.20.0.0/22` (1024 addresses).
The network is segmented into four VLANs:

| VLAN   | Name  | Max Hosts |
| ------ | ----- | --------- |
| VLAN10 | Sales | 200       |
| VLAN20 | IT    | 100       |
| VLAN30 | HR    | 50        |
| VLAN40 | Guest | 30        |

### Design Principles

**Broadcast domain separation:** Each VLAN maps to exactly one subnet. An
ARP broadcast in VLAN10 (Sales) does not reach VLAN20 (IT). This reduces
noise and limits the blast radius of broadcast storms.

**One subnet per VLAN / One VLAN per subnet:** Each subnet has a single Layer
3 gateway (a routed interface on the distribution switch or firewall). The
gateway IP is typically the first usable host address.

**Security boundary alignment:** Firewall ACLs and security policies are
applied at the Layer 3 boundary between subnets/VLANs. Placing HR on a
separate /26 makes it trivial to block direct access from Guest (VLAN40) to
HR (VLAN30) without affecting Sales→IT traffic.

### VLSM Allocation from `10.20.0.0/22`

Sort by host requirement (largest first), find minimum prefix:

```
200 hosts → /24 provides 254 (next power: 2^8=256, 256-2=254 ≥ 200)
100 hosts → /25 provides 126 (2^7=128, 128-2=126 ≥ 100)
 50 hosts → /26 provides  62 (2^6=64, 64-2=62 ≥ 50)
 30 hosts → /27 provides  30 (2^5=32, 32-2=30 ≥ 30)
```

Sequential allocation starting at `10.20.0.0`:

| VLAN   | Name  | CIDR           | Gateway     | Usable Range (first–last) | Broadcast   | Usable |
| ------ | ----- | -------------- | ----------- | ------------------------- | ----------- | ------ |
| VLAN10 | Sales | 10.20.0.0/24   | 10.20.0.1   | 10.20.0.1 – 10.20.0.254   | 10.20.0.255 | 254    |
| VLAN20 | IT    | 10.20.1.0/25   | 10.20.1.1   | 10.20.1.1 – 10.20.1.126   | 10.20.1.127 | 126    |
| VLAN30 | HR    | 10.20.1.128/26 | 10.20.1.129 | 10.20.1.129 – 10.20.1.190 | 10.20.1.191 | 62     |
| VLAN40 | Guest | 10.20.1.192/27 | 10.20.1.193 | 10.20.1.193 – 10.20.1.222 | 10.20.1.223 | 30     |

> **Note:** The "Usable Range" column shows all addresses that can be
> assigned to devices, including the gateway interface itself. The gateway
> takes the **first** address in each row. DHCP pools for client devices
> should therefore start at gateway + 1 (e.g. `10.20.0.2` for Sales,
> `10.20.1.2` for IT, `10.20.1.130` for HR, `10.20.1.194` for Guest).

Total allocated: 256 + 128 + 64 + 32 = **480 addresses**  
Available in /22: **1024 addresses**  
Remaining: **544 addresses** (for future VLANs)

### Gateway Assignment Pattern

The gateway address (the router's sub-interface) is traditionally assigned
the **first usable host address** in each subnet. For VLAN30 (HR):

```
Network:  10.20.1.128
Gateway:  10.20.1.129    ← L3 switch sub-interface for VLAN30
```

Static DHCP pool for VLAN30 would begin at `10.20.1.130` and end at
`10.20.1.190`, reserving `.129` for the gateway and leaving a few addresses
at the end for static assignments (printers, servers).

### Security Policy Example

```
ACL applied at VLAN boundary:
  permit VLAN10 (Sales)   → VLAN20 (IT)      tcp dst 443     # HTTPS to IT services
  permit VLAN20 (IT)      → any                               # IT full access
  deny   VLAN40 (Guest)   → VLAN30 (HR)      any             # Guests cannot reach HR
  deny   VLAN40 (Guest)   → VLAN10 (Sales)   any             # Guests cannot reach Sales
  permit VLAN40 (Guest)   → 0.0.0.0/0        tcp dst 80,443  # Internet only
```

---

## 9. Algorithm Pseudocode Reference

### 9.1 Subnet Split Algorithm (FLSM)

Divides a network block into `N` equal-sized subnets, each with prefix
`newPrefix`.

```
Input:
  network   *IPNet     — base network to split
  newPrefix int        — prefix length for each resulting subnet

Preconditions:
  newPrefix > network.prefix
  newPrefix <= addressBits (32 for IPv4, 128 for IPv6)

Algorithm:
  count     = 2^(newPrefix - network.prefix)   // number of subnets
  blockSize = 2^(addressBits - newPrefix)       // addresses per subnet

  result = []
  current = network.networkAddress              // start at network base
  mask    = cidrMask(newPrefix, addressBits)

  for i = 0 to count-1:
      subnet = IPNet{IP: current, Mask: mask}
      result.append(subnet)
      current = ipAddOffset(current, blockSize) // advance by block size

  return result

ipAddOffset(ip, offset):
  n = bigIntFromBytes(ip)
  n = n + offset
  return bigIntToBytes(n, len(ip))
```

**`netx` implementation:** `netx.Split(network, newPrefix)`

---

### 9.2 VLSM Allocation Algorithm

Allocates variable-length subnets from a base network to satisfy a list of
host-count requirements.

```
Input:
  base             *IPNet  — base network to allocate from
  hostRequirements []int   — number of usable hosts required per subnet

Algorithm:
  Sort hostRequirements in DESCENDING order  // largest-first allocation

  baseSize = bigIntFromBytes(base.IP)
  cursor   = baseSize                        // current allocation pointer
  endAddr  = baseSize + networkSize(base)    // exclusive end of base block

  result = []

  for each h in hostRequirements:
      p = prefixForHosts(h, addressBits)     // smallest prefix satisfying h

      if p == -1:
          return error("no valid prefix for {} hosts", h)

      blockSize = 2^(addressBits - p)
      newIP     = bigIntToBytes(cursor, addressLength)
      newMask   = cidrMask(p, addressBits)
      subnet    = buildSubnet(IPNet{IP: newIP, Mask: newMask})

      if cursor + blockSize > endAddr:
          return error("address space exhausted")

      result.append(subnet)
      cursor = cursor + blockSize            // advance past this subnet

  return result

prefixForHosts(hosts, bits):
  required = BigInt(hosts)
  for p = bits downto 0:                    // most-specific to least-specific
      usable = computeTotalHosts(p, bits)
      if usable >= required:
          return p
  return -1                                 // impossible

computeTotalHosts(p, bits):
  hostBits = bits - p
  if hostBits == 0: return 1               // /32 single host
  if hostBits == 1: return 2               // /31 point-to-point
  return 2^hostBits - 2                    // standard formula
```

**`netx` implementation:** `netx.DivideByHosts(base, hostRequirements)`

---

### 9.3 Route Summarization Algorithm

Aggregates a list of contiguous subnets into the shortest covering prefix.

```
Input:
  networks []*IPNet  — list of networks to summarize
                       (must be contiguous and start-aligned)

Algorithm:
  if len(networks) == 0: return nil, error

  // Convert all network addresses to big.Int
  addrs = [bigIntFromBytes(n.IP) for n in networks]

  // Find the minimum (lowest address) and maximum (highest address)
  minAddr = min(addrs)
  maxAddr = max(addrs)
  bits    = addressBits(networks[0])

  // Widen the prefix until the block covers minAddr..maxAddr
  // Start from the shortest prefix of the first address
  startPrefix = networks[0].prefix
  for p = startPrefix downto 0:
      mask     = cidrMask(p, bits)
      netAddr  = bigIntFromBytes(applyMask(minAddr, mask))  // AND
      bcast    = bigIntFromBytes(broadcastOf(netAddr, mask)) // OR with ^mask
      if netAddr == minAddr and bcast >= maxAddr:
          return IPNet{IP: bigIntToBytes(netAddr, addrLen), Mask: mask}

  return error("cannot summarize: networks are not contiguous")

// Validation: a valid summary must exactly cover all input networks.
// If bcast < maxAddr, the summary is too small.
// If netAddr != minAddr, the networks don't align to the proposed prefix.
```

---

### 9.4 Host Count and Prefix Search

Find the tightest prefix that satisfies a host requirement.

```
prefixForHosts(hosts, bits):
  // Walk from /bits (most specific, 1 host) to /0 (least specific)
  required = BigInt(hosts)
  for prefix = bits downto 0:
      usable = computeTotalHosts(prefix, bits)
      if usable >= required:
          return prefix       // first match is the tightest (smallest block)
  return -1
```

This scans from specific to general, so the first match is always the
smallest block satisfying the requirement — maximising address efficiency.

---

## 10. Mapping Concepts to the `netx` Go API

The `netx` library implements all the algorithms described above. The table
below maps each subnetting concept to its corresponding Go function.

### Parsing and Subnet Introspection

| Concept                | `netx` API                        | Returns           |
| ---------------------- | --------------------------------- | ----------------- |
| Parse CIDR notation    | `netx.ParseCIDR(cidr string)`     | `(Subnet, error)` |
| Parse (panic on error) | `netx.MustParseCIDR(cidr string)` | `Subnet`          |
| Network address        | `sub.NetworkAddress()`            | `net.IP`          |
| Broadcast address      | `sub.BroadcastAddress()`          | `net.IP`          |
| First usable host      | `sub.FirstHost()`                 | `net.IP`          |
| Last usable host       | `sub.LastHost()`                  | `net.IP`          |
| Usable host count      | `sub.TotalHosts()`                | `*big.Int`        |
| Prefix length          | `sub.Prefix()`                    | `int`             |
| CIDR string            | `sub.String()`                    | `string`          |
| Underlying \*net.IPNet | `sub.IPNet()`                     | `*net.IPNet`      |

### FLSM Operations

| Concept                  | `netx` API                                         | Returns                 |
| ------------------------ | -------------------------------------------------- | ----------------------- |
| Split into equal subnets | `netx.Split(network *net.IPNet, newPrefix int)`    | `([]*net.IPNet, error)` |
| Split into exactly N     | `netx.SplitIntoN(network *net.IPNet, n int)`       | `([]*net.IPNet, error)` |
| Next contiguous subnet   | `netx.NextSubnet(ipnet *net.IPNet, newPrefix int)` | `(*net.IPNet, error)`   |

### VLSM Operations

| Concept                         | `netx` API                                            | Returns             |
| ------------------------------- | ----------------------------------------------------- | ------------------- |
| Variable-length host allocation | `netx.DivideByHosts(base *net.IPNet, hostReqs []int)` | `([]Subnet, error)` |
| Min prefix for N hosts          | `netx.PrefixForHosts(hosts, bits int)`                | `int`               |
| Usable host count for prefix    | `netx.HostCount(prefix, bits int)`                    | `*big.Int`          |

### Subnet Utilities

| Concept                      | `netx` API                                         | Returns    |
| ---------------------------- | -------------------------------------------------- | ---------- |
| Containment check            | `netx.Contains(network *net.IPNet, ip net.IP)`     | `bool`     |
| Overlap detection            | `netx.Overlaps(netA, netB *net.IPNet)`             | `bool`     |
| Total address count          | `netx.NetworkSize(ipnet *net.IPNet)`               | `*big.Int` |
| Slice of CIDRs to strings    | `netx.SubnetsToStrings(nets []*net.IPNet)`         | `[]string` |
| Allocated subnets to strings | `netx.AllocatedSubnetsToStrings(subnets []Subnet)` | `[]string` |

### Complete Example: VLSM Allocation from `10.0.0.0/24`

```go
package main

import (
    "fmt"
    "net"

    "github.com/polarixa/replify/pkg/netx"
)

func main() {
    _, base, _ := net.ParseCIDR("10.0.0.0/24")

    // Allocate for: Sales=100, IT=20, HR=5, P2P-A=2, P2P-B=2
    subnets, err := netx.DivideByHosts(base, []int{100, 20, 5, 2, 2})
    if err != nil {
        panic(err)
    }

    labels := []string{"Sales", "IT", "HR", "P2P-A", "P2P-B"}
    for i, s := range subnets {
        fmt.Printf("%-8s  %-20s  first=%-15s  last=%-15s  hosts=%s\n",
            labels[i],
            s.String(),
            s.FirstHost(),
            s.LastHost(),
            s.TotalHosts(),
        )
    }
}
```

Expected output:

```
Sales     10.0.0.0/25           first=10.0.0.1         last=10.0.0.126       hosts=126
IT        10.0.0.128/27         first=10.0.0.129       last=10.0.0.158       hosts=30
HR        10.0.0.160/29         first=10.0.0.161       last=10.0.0.166       hosts=6
P2P-A     10.0.0.168/31         first=10.0.0.168       last=10.0.0.169       hosts=2
P2P-B     10.0.0.170/31         first=10.0.0.170       last=10.0.0.171       hosts=2
```

### Complete Example: FLSM Split of `192.168.10.0/24` into 4 Subnets

```go
package main

import (
    "fmt"
    "net"

    "github.com/polarixa/replify/pkg/netx"
)

func main() {
    _, base, _ := net.ParseCIDR("192.168.10.0/24")

    // Divide into 4 equal /26 subnets (24 + 2 borrowed bits = /26)
    subnets, err := netx.SplitIntoN(base, 4)
    if err != nil {
        panic(err)
    }

    branches := []string{"London", "Paris", "Tokyo", "Sydney"}
    for i, n := range subnets {
        s, _ := netx.ParseCIDR(n.String())
        fmt.Printf("%-8s  %-22s  hosts=%s\n",
            branches[i], s.String(), s.TotalHosts())
    }
}
```

Expected output:

```
London    192.168.10.0/26         hosts=62
Paris     192.168.10.64/26        hosts=62
Tokyo     192.168.10.128/26       hosts=62
Sydney    192.168.10.192/26       hosts=62
```

---

_This guide is part of the `netx` package documentation. For the Go API
reference, see [README.md](../README.md). For the cross-package architecture
and deduplication rationale, see the Cross-Package Architecture section in
the README._
