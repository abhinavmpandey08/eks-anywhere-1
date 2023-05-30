package ipam

import (
	"fmt"
	"net"

	"github.com/aws/eks-anywhere/pkg/networkutils"
)

type SequentialIPAM struct {
	cidr               *net.IPNet
	currentIP          net.IP
	firstIP            net.IP
	netClient          networkutils.NetClient
	isCurrentIPUpdated bool
}

type Option func(*SequentialIPAM)

func NewSequentialIPAM(networkCidr string, options ...Option) (IPAM, error) {
	ip, cidr, err := net.ParseCIDR(networkCidr)
	if err != nil {
		return nil, fmt.Errorf("parsing CIDR [%s]: %s", networkCidr, err)
	}
	ip = ip.Mask(cidr.Mask)

	ipam := &SequentialIPAM{
		cidr:               cidr,
		currentIP:          ip,
		firstIP:            ip,
		netClient:          &networkutils.DefaultNetClient{},
		isCurrentIPUpdated: false,
	}
	for _, opt := range options {
		opt(ipam)
	}

	return ipam, nil
}

func WithNetClient(netClient networkutils.NetClient) func(*SequentialIPAM) {
	return func(ipam *SequentialIPAM) {
		ipam.netClient = netClient
	}
}

func (ipam *SequentialIPAM) ReserveIPPool(count int) (networkutils.IPPool, error) {
	pool := networkutils.NewIPPool()
	for i := 0; i < count; i++ {
		if err := ipam.nextIP(); err != nil {
			return nil, err
		}
		pool.AddIP(ipam.currentIP.String())
	}
	return pool, nil
}

func (ipam *SequentialIPAM) nextIP() error {
	for {
		// condition to handle edge case where nextIP() is called for the very first time 
		if ipam.isCurrentIPUpdated || !ipam.currentIP.Equal(ipam.firstIP) {
			ipam.currentIP = incrementIP(ipam.currentIP)
		}
		ipam.isCurrentIPUpdated = true

		if !ipam.cidr.Contains(ipam.currentIP) {
			return fmt.Errorf("no more IPs available in CIDR [%s]", ipam.cidr.String())
		}

		// skip the first 3 and the last 2 IPs in the subnet as they are generally reserved
		if ipam.currentIP[3] <= 2 || ipam.currentIP[3] >= 254 {
			continue
		}

		// skip IPs that are already in use
		if networkutils.IsIPInUse(ipam.netClient, ipam.currentIP.String()) {
			continue
		}

		break
	}

	return nil
}

func incrementIP(ip net.IP) net.IP {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}
	return ip
}
