package iplookuptree

import (
	"encoding/binary"
	"math"
	"net"
)

// bitslen contstant must be a power of 2. It indicates how much space
// will be taken by a tree and maximum number of hops (treenode accesses). When bitslen is 4,
// the maximum number of hops will be 32 / bitslen and one node takes
// 1<< bitslen * (sizeof SID and *treenode). So current constant (4) will make maximum 8 hops and
// every node consumes 256 bytes.
const bitslen = 4

// SID type contains a list of corresponding service indexes.
// Every nth bit indicates nth service. So 0x1 stores (0),
// 0x2 stores (1), 0x3 stores (0, 1) services and etc.
// So you can handle up to 64 services.
// 0 indicates that there are no service.
// Example: service has index 6, then its SID representation will be 1<<6
type SID uint64

type Tree struct {
	root *treenode
}

type treenode struct {
	srvs [1<<bitslen]SID
	ptrs [1<<bitslen]*treenode
}

// isEmpty checks if node is empty. Used to determine whether to delete node
// Reminder: don't accidentally remove root node
func (node *treenode) isEmpty() bool {
	for i := 0; i < 1<<bitslen; i++ {
		if node.srvs[i] != 0 || node.ptrs[i] != nil {
			return false
		}
	}
	return true
}

// New creates new IP subnet tree. It only works with IP v4
func New() *Tree {
	tree := &Tree{&treenode{}}
	return tree
}

// Add adds ip subnet to the tree.
// service should contain only one true bit (anyway it works well with multiple bits).
// ipnet.IP must be of the length 4 (IP v4).
// It is up to you to handle this.
// This method does not compress subnets - if you put 1.1.1.1/24 and 1.1.1.1/23
// of the same service, it will hold both subnets.
func (tree *Tree) Add(service SID, ipnet net.IPNet) {
	node := tree.root

	prefixLen, _ := ipnet.Mask.Size()

	curLen := bitslen
	for i := 0; i < 32 / bitslen; i++ {
		if curLen >= prefixLen {

			start := getSubstring(ipnet.IP, uint8(i))
			end := start + (1<<uint(curLen - prefixLen)) - 1

			for j := start; j <= end; j++ {
				node.srvs[j] = node.srvs[j] | service
			}
			break
		}

		ind := getSubstring(ipnet.IP, uint8(i))
		if node.ptrs[ind] != nil {
			node = node.ptrs[ind]
		} else {
			node.ptrs[ind] = &treenode{}
			node = node.ptrs[ind]
		}
		curLen += bitslen
	}
}

// Get returns SID which corresponds to this ip v4
// ipv4 must be of length 4 and it is up to you to
// handle this.
func (tree *Tree) Get(ipv4 []byte) SID {
	var ans SID
	cur := tree.root

	for i := 0; i < 32 / bitslen; i++ {
		ind := getSubstring(ipv4, uint8(i))
		ans = ans | cur.srvs[ind]
		if cur = cur.ptrs[ind]; cur == nil {
			break
		}
	}

	return ans
}

// getSubstring is helper function that returns substring of
// bits placed in range [index * bitslen, index * bitslen + bitslen)
func getSubstring(ipv4 []byte, index uint8) uint32 {
	var ans = binary.BigEndian.Uint32(ipv4)
	ans = ans <<(bitslen * index)
	ans = ans >>(32 - bitslen)

	return ans
}

// Remove removes subnet from the tree. It works properly if you add
// and remove the same subnet. However, it will lead to undefined behaviour, if
// you remove subnet which was not added before.
func (tree *Tree) Remove(service SID, ipnet net.IPNet) {
	reversedService := math.MaxUint64 ^ service

	node := tree.root

	path := make([]*treenode, 32 / bitslen, 32 / bitslen)
	indpath := make([]uint32, 32 / bitslen, 32 / bitslen)
	pathlen := 0

	prefixLen, _ := ipnet.Mask.Size()

	curLen := bitslen
	for i := 0; i < 32 / bitslen; i++ {
		path[pathlen] = node
		pathlen++
		if curLen >= prefixLen {

			start := getSubstring(ipnet.IP, uint8(i))
			end := start + (1<<uint(curLen - prefixLen)) - 1

			for j := start; j <= end; j++ {
				node.srvs[j] = node.srvs[j] & reversedService
			}
			break
		}

		ind := getSubstring(ipnet.IP, uint8(i))
		indpath[pathlen] = ind

		if node.ptrs[ind] != nil {
			node = node.ptrs[ind]
		}
		curLen += bitslen
	}

	for i := pathlen - 1; i > 0; i-- {
		node = path[i]
		parent := path[i - 1]
		ind := indpath[i]

		if node.isEmpty() {
			parent.ptrs[ind] = nil
		}
	}
}