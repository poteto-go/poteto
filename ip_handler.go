package poteto

import (
	"fmt"
	"net"
	"strings"

	"github.com/poteto0/poteto/constant"
)

type ipHandler struct {
	trustIpRanges    []*net.IPNet
	isTrustPrivateIp bool
}

type IPHandler interface {
	SetIsTrustPrivateIP(flag bool)
	RegisterTrustIPRange(ranges *net.IPNet)
	CanTrust(ip net.IP) bool
	GetIPFromXFFHeader(ctx Context) (string, error)
	GetRemoteIP(ctx Context) (string, error)
}

func (iph *ipHandler) SetIsTrustPrivateIP(flag bool) {
	iph.isTrustPrivateIp = flag
}

func (iph *ipHandler) RegisterTrustIPRange(ranges *net.IPNet) {
	iph.trustIpRanges = append(iph.trustIpRanges, ranges)
}

func (iph *ipHandler) CanTrust(ip net.IP) bool {
	if iph.isTrustPrivateIp && ip.IsPrivate() {
		return true
	}

	for _, trustRanges := range iph.trustIpRanges {
		if trustRanges.Contains(ip) {
			return true
		}
	}
	return false
}

// first not trusted ip
// cause first app can exploit X-Forwarded-For
func (iph *ipHandler) GetIPFromXFFHeader(ctx Context) (string, error) {
	reqHeader := ctx.GetRequest().Header
	xffs := reqHeader[constant.HEADER_X_FORWARDED_FOR]
	if len(xffs) == 0 {
		return iph.GetRemoteIP(ctx)
	}

	ips := strings.Split(strings.Join(xffs, ","), ",")
	// check from right
	for i := len(ips) - 1; i >= 0; i-- {
		ips[i] = strings.TrimSpace(ips[i])
		ips[i] = strings.TrimPrefix(ips[i], "[")
		ips[i] = strings.TrimSuffix(ips[i], "]")
		ip := net.ParseIP(ips[i])

		if ip == nil {
			return iph.GetRemoteIP(ctx)
		}

		// return first not trusted ip
		if !iph.CanTrust(ip) {
			return ip.String(), nil
		}
	}
	return strings.TrimSpace(ips[0]), nil
}

func (iph *ipHandler) GetRemoteIP(ctx Context) (string, error) {
	ip, _, err := net.SplitHostPort(
		strings.TrimSpace(ctx.GetRequest().RemoteAddr),
	)

	if err != nil {
		fmt.Println("Error in GetRemoteIP", err)
		return "", err
	}

	return ip, nil
}