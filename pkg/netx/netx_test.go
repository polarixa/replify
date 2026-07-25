package netx_test

import (
	"math/big"
	"net"
	"testing"

	"github.com/polarixa/replify/pkg/netx"
)

// ///////////////////////////
// Section: CIDR parsing tests
// ///////////////////////////

func TestNetx_ParseCIDR_Valid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		cidr      string
		network   string
		broadcast string
		first     string
		last      string
		prefix    int
		hosts     int64
	}{
		{
			cidr:    "10.0.0.0/8",
			network: "10.0.0.0", broadcast: "10.255.255.255",
			first: "10.0.0.1", last: "10.255.255.254",
			prefix: 8, hosts: 16777214,
		},
		{
			cidr:    "192.168.1.0/24",
			network: "192.168.1.0", broadcast: "192.168.1.255",
			first: "192.168.1.1", last: "192.168.1.254",
			prefix: 24, hosts: 254,
		},
		{
			cidr:    "172.16.0.0/12",
			network: "172.16.0.0", broadcast: "172.31.255.255",
			first: "172.16.0.1", last: "172.31.255.254",
			prefix: 12, hosts: 1048574,
		},
		{
			cidr:    "10.0.0.0/30",
			network: "10.0.0.0", broadcast: "10.0.0.3",
			first: "10.0.0.1", last: "10.0.0.2",
			prefix: 30, hosts: 2,
		},
		// /31 — RFC 3021 point-to-point
		{
			cidr:    "10.0.0.0/31",
			network: "10.0.0.0", broadcast: "10.0.0.1",
			first: "10.0.0.0", last: "10.0.0.1",
			prefix: 31, hosts: 2,
		},
		// /32 — single host
		{
			cidr:    "10.0.0.1/32",
			network: "10.0.0.1", broadcast: "10.0.0.1",
			first: "10.0.0.1", last: "10.0.0.1",
			prefix: 32, hosts: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.cidr, func(t *testing.T) {
			sub, err := netx.ParseCIDR(tt.cidr)
			if err != nil {
				t.Fatalf("ParseCIDR(%q) error = %v", tt.cidr, err)
			}
			if got := sub.NetworkAddress().String(); got != tt.network {
				t.Errorf("NetworkAddress() = %q, want %q", got, tt.network)
			}
			if got := sub.BroadcastAddress().String(); got != tt.broadcast {
				t.Errorf("BroadcastAddress() = %q, want %q", got, tt.broadcast)
			}
			if got := sub.FirstHost().String(); got != tt.first {
				t.Errorf("FirstHost() = %q, want %q", got, tt.first)
			}
			if got := sub.LastHost().String(); got != tt.last {
				t.Errorf("LastHost() = %q, want %q", got, tt.last)
			}
			if got := sub.Prefix(); got != tt.prefix {
				t.Errorf("Prefix() = %d, want %d", got, tt.prefix)
			}
			if got := sub.TotalHosts().Int64(); got != tt.hosts {
				t.Errorf("TotalHosts() = %d, want %d", got, tt.hosts)
			}
		})
	}
}

func TestNetx_ParseCIDR_Invalid(t *testing.T) {
	t.Parallel()
	invalid := []string{
		"",
		"not-a-cidr",
		"192.168.1.1",
		"300.0.0.0/24",
		"192.168.1.0/33",
	}
	for _, cidr := range invalid {
		t.Run(cidr, func(t *testing.T) {
			_, err := netx.ParseCIDR(cidr)
			if err == nil {
				t.Errorf("ParseCIDR(%q) expected error, got nil", cidr)
			}
		})
	}
}

func TestNetx_MustParseCIDR_Panic(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParseCIDR with invalid CIDR should panic")
		}
	}()
	netx.MustParseCIDR("invalid")
}

func TestNetx_MustParseCIDR_Valid(t *testing.T) {
	t.Parallel()
	sub := netx.MustParseCIDR("10.0.0.0/8")
	if sub.Prefix() != 8 {
		t.Errorf("Prefix() = %d, want 8", sub.Prefix())
	}
}

// ///////////////////////////
// Section: IPv6 parsing tests
// ///////////////////////////

func TestNetx_ParseCIDR_IPv6(t *testing.T) {
	t.Parallel()
	sub, err := netx.ParseCIDR("2001:db8::/32")
	if err != nil {
		t.Fatalf("ParseCIDR IPv6 error = %v", err)
	}
	if sub.Prefix() != 32 {
		t.Errorf("Prefix() = %d, want 32", sub.Prefix())
	}
	if sub.NetworkAddress().String() != "2001:db8::" {
		t.Errorf("NetworkAddress() = %q, want 2001:db8::", sub.NetworkAddress())
	}
	// /32 has 2^96 usable hosts
	expected := new(big.Int)
	expected.Lsh(big.NewInt(1), 96)
	expected.Sub(expected, big.NewInt(2))
	if sub.TotalHosts().Cmp(expected) != 0 {
		t.Errorf("TotalHosts() = %s, want %s", sub.TotalHosts(), expected)
	}
}

func TestNetx_ParseCIDR_IPv6_128(t *testing.T) {
	t.Parallel()
	sub, err := netx.ParseCIDR("::1/128")
	if err != nil {
		t.Fatalf("ParseCIDR error = %v", err)
	}
	if sub.TotalHosts().Int64() != 1 {
		t.Errorf("TotalHosts() = %d, want 1 for /128", sub.TotalHosts().Int64())
	}
}

// ///////////////////////////
// Section: Subnet String()
// ///////////////////////////

func TestNetx_Subnet_String(t *testing.T) {
	t.Parallel()
	sub := netx.MustParseCIDR("10.1.0.0/16")
	if sub.String() != "10.1.0.0/16" {
		t.Errorf("String() = %q, want 10.1.0.0/16", sub.String())
	}
}

func TestNetx_Subnet_IPNet(t *testing.T) {
	t.Parallel()
	sub := netx.MustParseCIDR("10.0.0.0/24")
	if sub.IPNet() == nil {
		t.Error("IPNet() should not be nil")
	}
}

// ///////////////////////////
// Section: FLSM Split tests
// ///////////////////////////

func TestNetx_Split_IPv4_24_into_26(t *testing.T) {
	t.Parallel()
	base := netx.MustParseCIDR("10.0.0.0/24").IPNet()
	subnets, err := netx.Split(base, 26)
	if err != nil {
		t.Fatalf("Split error = %v", err)
	}
	want := []string{
		"10.0.0.0/26",
		"10.0.0.64/26",
		"10.0.0.128/26",
		"10.0.0.192/26",
	}
	if len(subnets) != len(want) {
		t.Fatalf("Split returned %d subnets, want %d", len(subnets), len(want))
	}
	for i, got := range netx.SubnetsToStrings(subnets) {
		if got != want[i] {
			t.Errorf("subnets[%d] = %q, want %q", i, got, want[i])
		}
	}
}

func TestNetx_Split_IPv4_16_into_24(t *testing.T) {
	t.Parallel()
	base := netx.MustParseCIDR("10.0.0.0/16").IPNet()
	subnets, err := netx.Split(base, 24)
	if err != nil {
		t.Fatalf("Split error = %v", err)
	}
	if len(subnets) != 256 {
		t.Errorf("Split into /24 from /16: got %d subnets, want 256", len(subnets))
	}
	if subnets[0].String() != "10.0.0.0/24" {
		t.Errorf("first subnet = %q, want 10.0.0.0/24", subnets[0])
	}
	if subnets[255].String() != "10.0.255.0/24" {
		t.Errorf("last subnet = %q, want 10.0.255.0/24", subnets[255])
	}
}

func TestNetx_Split_Error_NewPrefixNotLarger(t *testing.T) {
	t.Parallel()
	base := netx.MustParseCIDR("10.0.0.0/24").IPNet()
	_, err := netx.Split(base, 24)
	if err == nil {
		t.Error("Split with equal prefix should return error")
	}
	_, err = netx.Split(base, 20)
	if err == nil {
		t.Error("Split with smaller prefix should return error")
	}
}

func TestNetx_Split_Error_PrefixTooLarge(t *testing.T) {
	t.Parallel()
	base := netx.MustParseCIDR("10.0.0.0/24").IPNet()
	_, err := netx.Split(base, 33)
	if err == nil {
		t.Error("Split with prefix > 32 should return error")
	}
}

func TestNetx_Split_Error_NilNetwork(t *testing.T) {
	t.Parallel()
	_, err := netx.Split(nil, 26)
	if err == nil {
		t.Error("Split with nil network should return error")
	}
}

func TestNetx_SplitIntoN_4(t *testing.T) {
	t.Parallel()
	base := netx.MustParseCIDR("10.0.0.0/24").IPNet()
	subnets, err := netx.SplitIntoN(base, 4)
	if err != nil {
		t.Fatalf("SplitIntoN error = %v", err)
	}
	if len(subnets) != 4 {
		t.Errorf("SplitIntoN(4) returned %d subnets, want 4", len(subnets))
	}
}

func TestNetx_SplitIntoN_Error_NotPowerOfTwo(t *testing.T) {
	t.Parallel()
	base := netx.MustParseCIDR("10.0.0.0/24").IPNet()
	_, err := netx.SplitIntoN(base, 3)
	if err == nil {
		t.Error("SplitIntoN with non-power-of-2 should return error")
	}
}

func TestNetx_SplitIntoN_Error_LessThanTwo(t *testing.T) {
	t.Parallel()
	base := netx.MustParseCIDR("10.0.0.0/24").IPNet()
	_, err := netx.SplitIntoN(base, 1)
	if err == nil {
		t.Error("SplitIntoN with n=1 should return error")
	}
}

// ///////////////////////////
// Section: VLSM DivideByHosts tests
// ///////////////////////////

func TestNetx_DivideByHosts_Basic(t *testing.T) {
	t.Parallel()
	base := netx.MustParseCIDR("10.0.0.0/24").IPNet()
	subnets, err := netx.DivideByHosts(base, []int{100, 50, 10})
	if err != nil {
		t.Fatalf("DivideByHosts error = %v", err)
	}
	if len(subnets) != 3 {
		t.Fatalf("DivideByHosts returned %d subnets, want 3", len(subnets))
	}
	// /25 provides 126 hosts (≥ 100)
	if subnets[0].Prefix() != 25 {
		t.Errorf("subnets[0].Prefix() = %d, want 25", subnets[0].Prefix())
	}
	// /26 provides 62 hosts (≥ 50)
	if subnets[1].Prefix() != 26 {
		t.Errorf("subnets[1].Prefix() = %d, want 26", subnets[1].Prefix())
	}
	// /28 provides 14 hosts (≥ 10)
	if subnets[2].Prefix() != 28 {
		t.Errorf("subnets[2].Prefix() = %d, want 28", subnets[2].Prefix())
	}
	// Verify no overlap between allocated subnets.
	for i := 0; i < len(subnets); i++ {
		for j := i + 1; j < len(subnets); j++ {
			if netx.Overlaps(subnets[i].IPNet(), subnets[j].IPNet()) {
				t.Errorf("subnets[%d] and subnets[%d] overlap", i, j)
			}
		}
	}
}

func TestNetx_DivideByHosts_AllWithinBase(t *testing.T) {
	t.Parallel()
	base := netx.MustParseCIDR("10.0.0.0/24").IPNet()
	subnets, err := netx.DivideByHosts(base, []int{100, 50, 10})
	if err != nil {
		t.Fatalf("DivideByHosts error = %v", err)
	}
	for i, s := range subnets {
		if !base.Contains(s.NetworkAddress()) {
			t.Errorf("subnets[%d] network address %s outside base %s",
				i, s.NetworkAddress(), base)
		}
	}
}

func TestNetx_DivideByHosts_InsufficientSpace(t *testing.T) {
	t.Parallel()
	// A /30 only has 2 usable hosts; asking for 100 should fail.
	base := netx.MustParseCIDR("10.0.0.0/30").IPNet()
	_, err := netx.DivideByHosts(base, []int{100})
	if err == nil {
		t.Error("DivideByHosts with insufficient space should return error")
	}
}

func TestNetx_DivideByHosts_InvalidHostCount(t *testing.T) {
	t.Parallel()
	base := netx.MustParseCIDR("10.0.0.0/24").IPNet()
	_, err := netx.DivideByHosts(base, []int{0})
	if err == nil {
		t.Error("DivideByHosts with host count 0 should return error")
	}
}

func TestNetx_DivideByHosts_NilBase(t *testing.T) {
	t.Parallel()
	_, err := netx.DivideByHosts(nil, []int{10})
	if err == nil {
		t.Error("DivideByHosts with nil base should return error")
	}
}

func TestNetx_DivideByHosts_EmptyRequirements(t *testing.T) {
	t.Parallel()
	base := netx.MustParseCIDR("10.0.0.0/24").IPNet()
	_, err := netx.DivideByHosts(base, nil)
	if err == nil {
		t.Error("DivideByHosts with empty requirements should return error")
	}
}

func TestNetx_DivideByHosts_SortsDescending(t *testing.T) {
	t.Parallel()
	base := netx.MustParseCIDR("10.0.0.0/16").IPNet()
	// Pass requirements in ascending order; allocator must sort them.
	subnets, err := netx.DivideByHosts(base, []int{10, 50, 200})
	if err != nil {
		t.Fatalf("DivideByHosts error = %v", err)
	}
	// First allocated subnet should satisfy 200 hosts (largest requirement).
	if subnets[0].TotalHosts().Int64() < 200 {
		t.Errorf("first subnet hosts = %d; want ≥ 200", subnets[0].TotalHosts().Int64())
	}
}

func TestNetx_AllocatedSubnetsToStrings(t *testing.T) {
	t.Parallel()
	base := netx.MustParseCIDR("10.0.0.0/24").IPNet()
	subnets, _ := netx.DivideByHosts(base, []int{100, 50})
	strs := netx.AllocatedSubnetsToStrings(subnets)
	if len(strs) != 2 {
		t.Errorf("AllocatedSubnetsToStrings len = %d, want 2", len(strs))
	}
}

// ///////////////////////////
// Section: Contains tests
// ///////////////////////////

func TestNetx_Contains(t *testing.T) {
	t.Parallel()
	_, n, _ := net.ParseCIDR("10.0.0.0/8")
	tests := []struct {
		ip   string
		want bool
	}{
		{"10.0.0.1", true},
		{"10.255.255.255", true},
		{"10.0.0.0", true},
		{"11.0.0.1", false},
		{"192.168.1.1", false},
	}
	for _, tt := range tests {
		ip := net.ParseIP(tt.ip)
		if got := netx.Contains(n, ip); got != tt.want {
			t.Errorf("Contains(10.0.0.0/8, %s) = %v, want %v", tt.ip, got, tt.want)
		}
	}
}

func TestNetx_Contains_NilArgs(t *testing.T) {
	t.Parallel()
	_, n, _ := net.ParseCIDR("10.0.0.0/8")
	if netx.Contains(nil, net.ParseIP("10.0.0.1")) {
		t.Error("Contains(nil, ip) should return false")
	}
	if netx.Contains(n, nil) {
		t.Error("Contains(n, nil) should return false")
	}
}

// ///////////////////////////
// Section: Overlaps tests
// ///////////////////////////

func TestNetx_Overlaps(t *testing.T) {
	t.Parallel()
	tests := []struct {
		a, b string
		want bool
	}{
		{"10.0.0.0/24", "10.0.0.128/25", true},  // b inside a
		{"10.0.0.0/25", "10.0.0.128/25", false}, // adjacent, no overlap
		{"10.0.0.0/24", "10.0.1.0/24", false},   // different /24s
		{"10.0.0.0/8", "10.1.2.0/24", true},     // smaller inside larger
		{"192.168.0.0/16", "10.0.0.0/8", false}, // disjoint
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			subA := netx.MustParseCIDR(tt.a).IPNet()
			subB := netx.MustParseCIDR(tt.b).IPNet()
			if got := netx.Overlaps(subA, subB); got != tt.want {
				t.Errorf("Overlaps(%s, %s) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestNetx_Overlaps_NilArgs(t *testing.T) {
	t.Parallel()
	_, n, _ := net.ParseCIDR("10.0.0.0/8")
	if netx.Overlaps(nil, n) {
		t.Error("Overlaps(nil, n) should return false")
	}
	if netx.Overlaps(n, nil) {
		t.Error("Overlaps(n, nil) should return false")
	}
}

// ///////////////////////////
// Section: NetworkSize tests
// ///////////////////////////

func TestNetx_NetworkSize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		cidr string
		size int64
	}{
		{"10.0.0.0/24", 256},
		{"10.0.0.0/32", 1},
		{"10.0.0.0/31", 2},
		{"10.0.0.0/30", 4},
		{"10.0.0.0/8", 16777216},
	}
	for _, tt := range tests {
		t.Run(tt.cidr, func(t *testing.T) {
			_, n, _ := net.ParseCIDR(tt.cidr)
			got := netx.NetworkSize(n).Int64()
			if got != tt.size {
				t.Errorf("NetworkSize(%s) = %d, want %d", tt.cidr, got, tt.size)
			}
		})
	}
}

func TestNetx_NetworkSize_Nil(t *testing.T) {
	t.Parallel()
	if netx.NetworkSize(nil).Sign() != 0 {
		t.Error("NetworkSize(nil) should return 0")
	}
}

// ///////////////////////////
// Section: HostCount tests
// ///////////////////////////

func TestNetx_HostCount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		prefix, bits int
		want         int64
	}{
		{24, 32, 254},
		{30, 32, 2},
		{31, 32, 2},
		{32, 32, 1},
		{8, 32, 16777214},
	}
	for _, tt := range tests {
		got := netx.HostCount(tt.prefix, tt.bits).Int64()
		if got != tt.want {
			t.Errorf("HostCount(%d, %d) = %d, want %d", tt.prefix, tt.bits, got, tt.want)
		}
	}
}

// ///////////////////////////
// Section: PrefixForHosts tests
// ///////////////////////////

func TestNetx_PrefixForHosts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		hosts, bits int
		want        int
	}{
		{254, 32, 24},
		{100, 32, 25}, // /25 gives 126 hosts
		{50, 32, 26},  // /26 gives  62 hosts
		{10, 32, 28},  // /28 gives  14 hosts
		{2, 32, 31},   // /31 gives   2 hosts (RFC 3021 point-to-point)
		{1, 32, 32},   // /32 gives   1 host
		{255, 32, 23}, // /23 gives 510 hosts
	}
	for _, tt := range tests {
		got := netx.PrefixForHosts(tt.hosts, tt.bits)
		if got != tt.want {
			t.Errorf("PrefixForHosts(%d, %d) = %d, want %d",
				tt.hosts, tt.bits, got, tt.want)
		}
	}
}

// ///////////////////////////
// Section: NextSubnet tests
// ///////////////////////////

func TestNetx_NextSubnet(t *testing.T) {
	t.Parallel()
	tests := []struct {
		cidr      string
		newPrefix int
		want      string
	}{
		{"10.0.0.0/26", 26, "10.0.0.64/26"},
		{"10.0.0.64/26", 26, "10.0.0.128/26"},
		{"10.0.0.192/26", 26, "10.0.1.0/26"},
		{"192.168.0.0/24", 24, "192.168.1.0/24"},
		{"10.0.0.0/30", 30, "10.0.0.4/30"},
	}
	for _, tt := range tests {
		t.Run(tt.cidr, func(t *testing.T) {
			base := netx.MustParseCIDR(tt.cidr).IPNet()
			next, err := netx.NextSubnet(base, tt.newPrefix)
			if err != nil {
				t.Fatalf("NextSubnet error = %v", err)
			}
			if got := next.String(); got != tt.want {
				t.Errorf("NextSubnet(%s, %d) = %q, want %q",
					tt.cidr, tt.newPrefix, got, tt.want)
			}
		})
	}
}

func TestNetx_NextSubnet_NilInput(t *testing.T) {
	t.Parallel()
	_, err := netx.NextSubnet(nil, 24)
	if err == nil {
		t.Error("NextSubnet(nil, 24) should return error")
	}
}

func TestNetx_NextSubnet_InvalidPrefix(t *testing.T) {
	t.Parallel()
	base := netx.MustParseCIDR("10.0.0.0/24").IPNet()
	_, err := netx.NextSubnet(base, 33)
	if err == nil {
		t.Error("NextSubnet with prefix > 32 should return error")
	}
}

// ///////////////////////////
// Section: SubnetsToStrings tests
// ///////////////////////////

func TestNetx_SubnetsToStrings(t *testing.T) {
	t.Parallel()
	base := netx.MustParseCIDR("10.0.0.0/24").IPNet()
	nets, _ := netx.Split(base, 26)
	strs := netx.SubnetsToStrings(nets)
	want := []string{"10.0.0.0/26", "10.0.0.64/26", "10.0.0.128/26", "10.0.0.192/26"}
	for i, s := range strs {
		if s != want[i] {
			t.Errorf("SubnetsToStrings[%d] = %q, want %q", i, s, want[i])
		}
	}
}
