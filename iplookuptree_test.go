package iplookuptree

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"testing"
)


type ipNetList []net.IPNet

func newIPNetList(rawCIDRs []string) (ipNetList, error) {
	ans := make([]net.IPNet, len(rawCIDRs))

	for i, rawCIDR := range rawCIDRs {
		_, ipv4Net, err := net.ParseCIDR(rawCIDR)

		if err != nil {

			return nil, fmt.Errorf("could not parse CIDR at %d with cidr %s", i, rawCIDR)
		}
		ans[i] = *ipv4Net
	}

	return ans, nil
}

func newIPNetListFromFile(filename string) (cidrList ipNetList, err error) {
	f, err := os.Open(filename)

	if err != nil {
		return
	}

	defer f.Close()

	r := bufio.NewScanner(f)

	var lines []string

	for r.Scan() {
		lines = append(lines, r.Text())
	}

	if err = r.Err(); err != nil {
		return
	}

	return newIPNetList(lines)
}

func Test_getSubstring(t *testing.T) {
	// 00001010 11000000 10101011 00011000
	ipv4bytes := []byte{0x0A, 0xC0, 0xAB, 0x18}

	if found := getSubstring(ipv4bytes, 0); found != 0x0 {
		t.Errorf("getSubstring() - found %v", found)
	}

	if found := getSubstring(ipv4bytes, 1); found != 0xA {
		t.Errorf("getSubstring() - found %v", found)
	}

	if found := getSubstring(ipv4bytes, 2); found != 0xC {
		t.Errorf("getSubstring() - found %v", found)
	}

	if found := getSubstring(ipv4bytes, 3); found != 0x0 {
		t.Errorf("getSubstring() - found %v", found)
	}

	if found := getSubstring(ipv4bytes, 4); found != 0xA {
		t.Errorf("getSubstring() - found %v", found)
	}

	if found := getSubstring(ipv4bytes, 5); found != 0xB {
		t.Errorf("getSubstring() - found %v", found)
	}
	if found := getSubstring(ipv4bytes, 6); found != 0x1 {
		t.Errorf("getSubstring() - found %v", found)
	}

	if found := getSubstring(ipv4bytes, 7); found != 0x8 {
		t.Errorf("getSubstring() - found %v", found)
	}
}

func TestTree_Add(t *testing.T) {
	// 00001010 11000000 10101011 00011000
	ipv4bytes := []byte{0x0A, 0xC0, 0xAB, 0x18}

	ipnet := net.IPNet{ipv4bytes, []byte{0xFF, 0xFF, 0xFF, 0xFF}}

	tree := New()

	var id SID = 1

	tree.Add(id, ipnet)

	cur := tree.root
	index := 0x0
	if cur.ptrs[index] == nil {
		t.Errorf("Add() - expected to have pointer at %v", index)
	}

	cur = cur.ptrs[index]
	index = 0xA
	if cur.ptrs[index] == nil {
		t.Errorf("Add() - expected to have pointer at %v", index)
	}

	cur = cur.ptrs[index]
	index = 0xC
	if cur.ptrs[index] == nil {
		t.Errorf("Add() - expected to have pointer at %v", index)
	}

	cur = cur.ptrs[index]
	index = 0x0
	if cur.ptrs[index] == nil {
		t.Errorf("Add() - expected to have pointer at %v", index)
	}

	cur = cur.ptrs[index]
	index = 0xA
	if cur.ptrs[index] == nil {
		t.Errorf("Add() - expected to have pointer at %v", index)
	}

	cur = cur.ptrs[index]
	index = 0xB
	if cur.ptrs[index] == nil {
		t.Errorf("Add() - expected to have pointer at %v", index)
	}

	cur = cur.ptrs[index]
	index = 0x1
	if cur.ptrs[index] == nil {
		t.Errorf("Add() - expected to have pointer at %v", index)
	}

	cur = cur.ptrs[index]
	index = 0x8
	if cur.ptrs[index] != nil {
		t.Errorf("Add() - expected to not have pointer at %v", index)
	}


	if cur.srvs[index] | id != 1 {
		t.Errorf("Add() - expected to have service id, but found %v", cur.srvs[index])
	}
}

func TestTree_Remove(t *testing.T) {
	var id SID = 1
	tree := New()

	ipv4bytes := []byte{0x0A, 0xC0, 0xAB, 0x18}
	ipnet := net.IPNet{ipv4bytes, []byte{0xFF, 0xFF, 0xFF, 0xFF}}

	tree.Add(id, ipnet)
	tree.Remove(id, net.IPNet{[]byte{0x0A, 0xC0, 0xAB, 0x19}, []byte{0xFF, 0xFF, 0xFF, 0xFF}})

	if ans := tree.Get(ipv4bytes); ans != id {
		t.Errorf("Remove() - expected to find id")
	}

	tree.Remove(id, ipnet)

	if ans := tree.Get(ipv4bytes); ans != 0 {
		t.Errorf("Remove() - expected to find zero services")
	}

	if tree.root.isEmpty() != true {
		t.Errorf("Remove() - expected to not have any nodes")
	}
}

func TestTree_Remove2(t *testing.T) {
	var id1 SID = 1
	var id2 SID = 2
	tree := New()

	ipv4bytes := []byte{0x0A, 0xC0, 0xAB, 0x18}
	ipnet := net.IPNet{ipv4bytes, []byte{0xFF, 0xFF, 0xFF, 0xFF}}

	tree.Add(id1, ipnet)

	tree.Remove(id2, ipnet)

	if ans := tree.Get(ipv4bytes); ans != id1 {
		t.Errorf("Remove() - expected to find id1")
	}
}

func TestTree_Remove3(t *testing.T) {
	var id SID = 16
	tree := New()

	ipv4bytes1 := []byte{0x0A, 0xC0, 0xAB, 0x18}
	ipnet1 := net.IPNet{ipv4bytes1, []byte{0xFF, 0xFF, 0xFF, 0xFF}}

	ipv4bytes2 := []byte{0x0A, 0xC0, 0xAD, 0x18}
	ipnet2 := net.IPNet{ipv4bytes2, []byte{0xFF, 0xFF, 0xFF, 0xFF}}

	tree.Add(id, ipnet1)
	tree.Add(id, ipnet2)

	if ans := tree.Get(ipv4bytes1); ans != id {
		t.Errorf("Remove() - expected to find id")
	}

	if ans := tree.Get(ipv4bytes2); ans != id {
		t.Errorf("Remove() - expected to find id")
	}

	tree.Remove(id, ipnet1)

	if ans := tree.Get(ipv4bytes1); ans != 0 {
		t.Errorf("Remove() - expected to find zero services")
	}

	if ans := tree.Get(ipv4bytes2); ans != id {
		t.Errorf("Remove() - expected to find id")
	}
}

func TestTree_Get(t *testing.T) {
	t.Skip()
	var id SID = 1
	tree := New()

	ipNetList, _ := newIPNetListFromFile("./test_data/cidr_list.txt")

	for _, ipnet := range ipNetList {
		tree.Add(id, ipnet)
	}

	var ip []byte

	ip = net.ParseIP("31.13.64.51").To4()

	if found := tree.Get(ip); found != id {
		t.Errorf("Treenode_Get() - found %v", found)
	}

	ip = net.ParseIP("31.13.64.52").To4()

	if found := tree.Get(ip); found != 0 {
		t.Errorf("Treenode_Get() - found %v", found)
	}

	ip = net.ParseIP("198.11.251.32").To4()

	if found := tree.Get(ip); found != 1 {
		t.Errorf("Treenode_Get() - found %v", found)
	}

	ip = net.ParseIP("198.11.251.255").To4()

	if found := tree.Get(ip); found != 0 {
		t.Errorf("Treenode_Get() - found %v", found)
	}

	ip = net.ParseIP("217.76.74.226").To4()

	if found := tree.Get(ip); found != 1 {
		t.Errorf("Treenode_Get() - found %v", found)
	}
}



func BenchmarkTreenode_Get(b *testing.B) {
	b.Skip()
	const querynum = 1000 * 1000
	var id SID = 1
	tree := New()

	ipNetList, _ := newIPNetListFromFile("./test_data/cidr_list.txt")

	for _, ipnet := range ipNetList {
		tree.Add(id, ipnet)
	}

	b.ResetTimer()

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ip := make([]byte, 4)

		for j := 0; j < querynum; j++ {
			rand.Read(ip)
			tree.Get(ip)
		}
	}
}