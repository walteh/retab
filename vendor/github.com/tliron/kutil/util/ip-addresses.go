package util

import (
	"fmt"
	"net"
	"net/netip"
	"strings"
)

func JoinIPAddressPort(address string, port int) string {
	if IsIPv6(address) {
		return fmt.Sprintf("[%s]:%d", address, port)
	} else {
		return fmt.Sprintf("%s:%d", address, port)
	}
}

func IsIPv6(address string) bool {
	// See: https://stackoverflow.com/questions/22751035/golang-distinguish-ipv4-ipv6
	return strings.Contains(address, ":")
}

func IsUDPAddrEqual(a *net.UDPAddr, b *net.UDPAddr) bool {
	return a.IP.Equal(b.IP) && (a.Port == b.Port) && (a.Zone == b.Zone)
}

func ToReachableIPAddress(address string) (string, string, error) {
	if net.ParseIP(address).IsUnspecified() {
		isIpv6 := IsIPv6(address)

		if interfaces, err := net.Interfaces(); err == nil {
			// Try to find a global unicast first
			for _, interface_ := range interfaces {
				if (interface_.Flags&net.FlagLoopback == 0) && (interface_.Flags&net.FlagUp != 0) {
					if addrs, err := interface_.Addrs(); err == nil {
						for _, addr := range addrs {
							if addr_, ok := addr.(*net.IPNet); ok {
								//util.DumpIPAddress(addr_.IP.String())
								if addr_.IP.IsGlobalUnicast() {
									ip := addr_.IP.String()
									if isIpv6 == IsIPv6(ip) {
										return ip, "", nil
									}
								}
							}
						}
					} else {
						return "", "", err
					}
				}
			}

			// No global unicast available
			for _, interface_ := range interfaces {
				if (interface_.Flags&net.FlagLoopback == 0) && (interface_.Flags&net.FlagUp != 0) {
					if addrs, err := interface_.Addrs(); err == nil {
						for _, addr := range addrs {
							if addr_, ok := addr.(*net.IPNet); ok {
								ip := addr_.IP.String()
								if isIpv6 == IsIPv6(ip) {
									// The zone (required when not global unicast) is the interface name
									return ip, interface_.Name, nil
								}
							}
						}
					} else {
						return "", "", err
					}
				}
			}

			return "", "", fmt.Errorf("cannot find an equivalent reachable address for: %s", address)
		} else {
			return "", "", err
		}
	}

	return address, "", nil
}

func ToBroadcastIPAddress(address string) (string, string, error) {
	// Note: net.ParseIP can't parse IPv6 zone
	if ip, err := netip.ParseAddr(address); err == nil {
		if !ip.IsMulticast() {
			return "", "", fmt.Errorf("not a multicast address: %s", address)
		}

		if IsIPv6(address) && ip.Zone() == "" {
			if interfaces, err := net.Interfaces(); err == nil {
				for _, interface_ := range interfaces {
					//fmt.Printf("%s\n", interface_.Flags.String())
					if (interface_.Flags&net.FlagLoopback == 0) && (interface_.Flags&net.FlagUp != 0) &&
						(interface_.Flags&net.FlagBroadcast != 0) && (interface_.Flags&net.FlagMulticast != 0) {
						// The zone is the interface name
						return address, interface_.Name, nil
					}
				}
			} else {
				return "", "", err
			}

			return "", "", fmt.Errorf("cannot find a zone for: %s", address)
		}

		return address, "", nil
	} else {
		return "", "", err
	}
}

func DumpIPAddress(address any) {
	// Note: net.ParseIP can't parse IPv6 zone
	ip := netip.MustParseAddr(ToString(address))
	fmt.Printf("address: %s\n", ip)
	fmt.Printf("  global unicast:            %t\n", ip.IsGlobalUnicast())
	fmt.Printf("  interface local multicast: %t\n", ip.IsInterfaceLocalMulticast())
	fmt.Printf("  link local multicast:      %t\n", ip.IsLinkLocalMulticast())
	fmt.Printf("  link local unicast:        %t\n", ip.IsLinkLocalUnicast())
	fmt.Printf("  loopback:                  %t\n", ip.IsLoopback())
	fmt.Printf("  multicast:                 %t\n", ip.IsMulticast())
	fmt.Printf("  private:                   %t\n", ip.IsPrivate())
	fmt.Printf("  unspecified:               %t\n", ip.IsUnspecified())
}
