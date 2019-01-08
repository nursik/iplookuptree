# iplookuptree

### Overview
This package provides a data structure to make efficient IP lookups in the set of CIDRs.
It has zero dependencies and very fast lookups.
Benchmark shows 38 ns per lookup (half of this time is spent to generate random ip) for test data (261 CIDRs) and 103KB memory usage.

This package can be used with `gopacket` package to efficiently filter packets based on the CIDR list.
You can handle up to 64 filters (services).

`Add` and `Remove` methods are atomic.

### Usage

Simple usage
```go
package main

import (
	"github.com/nursik/iplookuptree"
	"fmt"
    "net"
)
func main() { 
    // You can only add ipv4 subnets to the tree
    _, ipv4Net, _ := net.ParseCIDR("10.10.10.10/24") 
	
    // Service index (between 0 and 63)
    var uint serviceIndex = 4
    var serviceID iplookuptree.SID = 1<<serviceIndex
    tree := iplookuptree.New()

    // Add subnet with corresponding service.
    tree.Add(serviceID, *ipv4Net)
    
    sid := tree.Get([]byte{0xA, 0x0A, 0x0A, 0x40})
    
    // Prints 16
    fmt.Println(sid)
    
    tree.Remove(serviceID, *ipv4Net)
    
    sid = tree.Get([]byte{0xA, 0x0A, 0x0A, 0x40})
    
    // Prints 0
    fmt.Println(sid)
}
```


### Cavets
#### Invalid input
As you may noticed, tree does not return any error. It is up to you to provide valid input.

You MUST always provide subnet of IPv4 (in `Add` and `Remove` methods), service must not be 0,
and ip must be of the length 4 (in `Get` method)

#### Remove method
Remove method does not work as you expect. If you `Add` "10.10.10.10/x" and `Remove` "10.10.10.10/x"
of the same service, it works properly as you expect. BUT if you `Add` "10.10.10.10/x" and `Remove` "10.10.10.10/y",
it remove subnet or do nothing (in case y > x) OR do nothing or leave gaps (in case y < x).

#### Subnets compression
Tree does not compress CIDRs list. So if there are "10.10.10.0/32" and "10.10.10.1/32", they will be treated as different
subnets (not as "10.10.10.0/31")

#### Memory usage
Every treenode comsumes 256 bytes. Every subnet is represented by 1 to 8 nodes (depending on subnet mask).
On the test data (261 CIDR list) consumes 108 KB. Profile your app for memory consumption.