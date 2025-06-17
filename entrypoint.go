package main

import (
	"os"
	"os/exec"
	"time"
)

func main() {
	dnscrypt := exec.Command("/usr/sbin/dnscrypt-proxy", "-config", "/etc/dnscrypt-proxy/dnscrypt-proxy.toml")
	dnscrypt.Stdout = os.Stdout
	dnscrypt.Stderr = os.Stderr

	if err := dnscrypt.Start(); err != nil {
		panic("dnscrypt-proxy failed: " + err.Error())
	}

	time.Sleep(1 * time.Second)

	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{
			"--keep-in-foreground",
			"--cache-size=1024",
			"--no-poll",
			"--no-resolv",
			"--server=127.0.0.1#5300",
		}
	}

	cmd := exec.Command("/usr/sbin/dnsmasq", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		panic("dnsmasq failed: " + err.Error())
	}
}
