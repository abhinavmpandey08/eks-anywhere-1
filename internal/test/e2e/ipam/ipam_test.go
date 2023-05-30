package ipam_test

import (
	"errors"
	"net"
	"testing"
	"time"

	. "github.com/aws/eks-anywhere/internal/test/e2e/ipam"

	. "github.com/onsi/gomega"
)

func TestSequentialIPAMReserveIPPoolInvalidCIDR(t *testing.T) {
	g := NewWithT(t)

	ipam, err := NewSequentialIPAM("10.0.0.0/64", WithNetClient(&MockNetClient{}))
	g.Expect(err).To(MatchError("parsing CIDR [10.0.0.0/64]: invalid CIDR address: 10.0.0.0/64"))
	g.Expect(ipam).To(BeNil())
}

func TestSequentialIPAMReserveIPPool(t *testing.T) {
	g := NewWithT(t)

	ipam, err := NewSequentialIPAM("10.0.0.0/23", WithNetClient(&MockNetClient{}))
	g.Expect(err).ToNot(HaveOccurred())

	pool, err := ipam.ReserveIPPool(252)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(pool.ToString()).To(ContainSubstring("10.0.0.253"))
	g.Expect(pool.ToString()).To(ContainSubstring("10.0.1.3"))

	g.Expect(pool.ToString()).ToNot(ContainSubstring("10.0.0.254"))
	g.Expect(pool.ToString()).ToNot(ContainSubstring("10.0.0.255"))
	g.Expect(pool.ToString()).ToNot(ContainSubstring("10.0.1.0"))
	g.Expect(pool.ToString()).ToNot(ContainSubstring("10.0.1.1"))
	g.Expect(pool.ToString()).ToNot(ContainSubstring("10.0.1.2"))
}

func TestSequentialIPAMReserveIPPoolOverflow(t *testing.T) {
	g := NewWithT(t)

	ipam, err := NewSequentialIPAM("10.0.0.0/24", WithNetClient(&MockNetClient{}))
	g.Expect(err).ToNot(HaveOccurred())

	pool, err := ipam.ReserveIPPool(251)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(pool.ToString()).To(ContainSubstring("10.0.0.253"))
	g.Expect(pool.ToString()).ToNot(ContainSubstring("10.0.1.3"))

	pool, err = ipam.ReserveIPPool(1)
	g.Expect(err).To(MatchError("no more IPs available in CIDR [10.0.0.0/24]"))
}

func TestSequentialIPAMReserveIPPoolOneIPSuccess(t *testing.T) {
	g := NewWithT(t)

	ipam, err := NewSequentialIPAM("192.168.0.20/32", WithNetClient(&MockNetClient{}))
	g.Expect(err).ToNot(HaveOccurred())

	pool, err := ipam.ReserveIPPool(1)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(pool.ToString()).To(ContainSubstring("192.168.0.20"))
}

func TestSequentialIPAMReserveIPPoolIPInUse(t *testing.T) {
	g := NewWithT(t)

	ipam, err := NewSequentialIPAM("192.168.0.10/32", WithNetClient(&MockNetClient{}))
	g.Expect(err).ToNot(HaveOccurred())

	_, err = ipam.ReserveIPPool(1)
	g.Expect(err).To(MatchError("no more IPs available in CIDR [192.168.0.10/32]"))
}

type MockNetClient struct{}

func (n *MockNetClient) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	// add dummy case for coverage
	if address == "192.168.0.10:80" {
		return &net.IPConn{}, nil
	}
	return nil, errors.New("")
}
