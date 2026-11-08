package engines

import (
	"fmt"

	"github.com/go-rod/rod/lib/launcher"

	"github.com/FacileStudio/sonar/internal/netguard"
)

type rodGuard struct {
	proxy *netguard.Proxy
}

func startGuard(g *netguard.Guard) (*rodGuard, error) {
	if g == nil {
		return &rodGuard{}, nil
	}
	p, err := g.StartProxy()
	if err != nil {
		return nil, fmt.Errorf("%w: guard proxy: %s", ErrTransient, err)
	}
	return &rodGuard{proxy: p}, nil
}

func (rg *rodGuard) addr() string {
	if rg.proxy == nil {
		return ""
	}
	return rg.proxy.Addr()
}

func (rg *rodGuard) close() {
	if rg.proxy != nil {
		_ = rg.proxy.Close()
	}
}

func (rg *rodGuard) verdict() error {
	if rg.proxy == nil || rg.proxy.Blocked() == 0 {
		return nil
	}
	return &netguard.BlockedError{Code: netguard.CodeInternal, Reason: "the page requested an internal address"}
}

func withProxy(l *launcher.Launcher, proxy string) *launcher.Launcher {
	if proxy == "" {
		return l
	}
	return l.
		Set("proxy-server", "http://"+proxy).
		Set("proxy-bypass-list", "<-loopback>").
		Set("host-resolver-rules", "MAP * ~NOTFOUND , EXCLUDE 127.0.0.1").
		Set("disable-quic").
		Set("disable-background-networking").
		Set("force-webrtc-ip-handling-policy", "disable_non_proxied_udp")
}
