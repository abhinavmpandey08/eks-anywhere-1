package ipam

import (
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/go-logr/logr"
)

type RandomIPAM struct {
	networkCidr string
	networkIPs  map[string]bool
	logger      logr.Logger
}

func NewRandomIPAM(logger logr.Logger, networkCidr string) IPAM {
	return &RandomIPAM{
		networkCidr: networkCidr,
		networkIPs:  make(map[string]bool),
		logger:      logger,
	}
}

func (ipam *RandomIPAM) ReserveIP() string {
	return ipam.getUniqueIP(ipam.networkCidr, ipam.networkIPs)
}

func (ipam *RandomIPAM) ReserveIPPool(count int) (networkutils.IPPool, error) {
	pool := networkutils.NewIPPool()
	for i := 0; i < count; i++ {
		pool.AddIP(ipam.ReserveIP())
	}
	return pool, nil
}

func (ipam *RandomIPAM) getUniqueIP(cidr string, usedIPs map[string]bool) string {
	ipgen := networkutils.NewIPGenerator(&networkutils.DefaultNetClient{})
	ip, err := ipgen.GenerateUniqueIP(cidr)
	for ; err != nil || usedIPs[ip]; ip, err = ipgen.GenerateUniqueIP(cidr) {
		if err != nil {
			ipam.logger.V(2).Info("Warning: getting unique IP failed", "error", err)
		}
		if usedIPs[ip] {
			ipam.logger.V(2).Info("Warning: generated IP is already taken", "IP", ip)
		}
	}
	usedIPs[ip] = true
	return ip
}
