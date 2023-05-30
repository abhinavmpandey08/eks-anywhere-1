package ipam

import (
	"github.com/aws/eks-anywhere/pkg/networkutils"
)

type IPAM interface {
	ReserveIPPool(count int) (networkutils.IPPool, error)
}
