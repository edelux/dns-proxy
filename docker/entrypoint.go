package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"syscall"
	"time"
)

const (
	dnsmasqConfPath    = "/etc/dnsmasq.conf"
	dnscryptConfigPath = "/etc/dnscrypt-proxy/dnscrypt-proxy.toml"
)

func main() {
	// Fix file permissions for services running as nobody user
	if err := fixFilePermissions(); err != nil {
		log.Printf("Warning: could not fix file permissions: %v", err)
	}
	
	// Process command line arguments
	args := os.Args[1:]
	
	for _, arg := range args {
		processArgument(arg)
	}
	
	// Start dnscrypt-proxy in background
	if err := startDnscryptProxy(); err != nil {
		log.Fatalf("Error starting dnscrypt-proxy: %v", err)
	}
	
	// Start dnsmasq in foreground
	if err := startDnsmasq(); err != nil {
		log.Fatalf("Error starting dnsmasq: %v", err)
	}
}

func processArgument(arg string) {
	// Regex for server parameters with IP validation (4 octets: 1-3 digits each)
	// Matches patterns like: server=/domain/192.168.1.1 or server=192.168.1.1
	serverRegex := regexp.MustCompile(`^-{0,2}(server=)(.*/)?(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})$`)
	
	// Regex for doh-server parameters
	dohServerRegex := regexp.MustCompile(`^-{0,2}(doh-server=)(.+)$`)
	
	if matches := serverRegex.FindStringSubmatch(arg); len(matches) == 4 {
		// Process server parameter - reconstruct the full server value
		domain := matches[2] // Could be empty or "/domain/"
		ip := matches[3]     // The validated IP
		var serverValue string
		if domain != "" {
			serverValue = domain + ip
		} else {
			serverValue = ip
		}
		
		if err := addServerToDnsmasq(serverValue); err != nil {
			log.Printf("Error adding server to dnsmasq.conf: %v", err)
		}
		log.Printf("Added server to dnsmasq.conf: %s (IP: %s)", serverValue, ip)
	} else if matches := dohServerRegex.FindStringSubmatch(arg); len(matches) == 3 {
		// Process doh-server parameter
		dohServerValue := matches[2]
		if err := updateDohServer(dohServerValue); err != nil {
			log.Printf("Error updating doh-server in dnscrypt-proxy.toml: %v", err)
		}
		log.Printf("Updated doh-server in dnscrypt-proxy.toml: %s", dohServerValue)
	} else {
		log.Printf("Invalid argument format: %s", arg)
	}
}

func addServerToDnsmasq(serverValue string) error {
	// Check if file exists and get its current permissions
	fileInfo, err := os.Stat(dnsmasqConfPath)
	needsPermissionFix := false
	if err == nil && fileInfo.Mode().Perm() != 0644 {
		needsPermissionFix = true
	}
	
	// Open file for appending
	file, err := os.OpenFile(dnsmasqConfPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open dnsmasq.conf: %v", err)
	}
	defer file.Close()
	
	// Add server parameter to the end of the file
	serverLine := fmt.Sprintf("server=%s\n", serverValue)
	if _, err := file.WriteString(serverLine); err != nil {
		return fmt.Errorf("could not write to dnsmasq.conf: %v", err)
	}
	
	// Fix permissions if needed
	if needsPermissionFix {
		if err := os.Chmod(dnsmasqConfPath, 0644); err != nil {
			log.Printf("Warning: could not fix permissions for dnsmasq.conf: %v", err)
		}
	}
	
	return nil
}

func updateDohServer(dohServerValue string) error {
	// Read dnscrypt-proxy.toml file
	content, err := ioutil.ReadFile(dnscryptConfigPath)
	if err != nil {
		return fmt.Errorf("could not read dnscrypt-proxy.toml: %v", err)
	}
	
	// Regex to find server_names = ['value'] line
	serverNamesRegex := regexp.MustCompile(`(server_names\s*=\s*\[')([^']+)('\])`)
	
	// Replace value between single quotes
	newContent := serverNamesRegex.ReplaceAllString(string(content), "${1}"+dohServerValue+"${3}")
	
	// Write modified content back to file with proper permissions (644)
	if err := ioutil.WriteFile(dnscryptConfigPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("could not write dnscrypt-proxy.toml: %v", err)
	}
	
	return nil
}

func startDnscryptProxy() error {
	log.Println("Starting dnscrypt-proxy in background...")
	
	// Verify dnscrypt-proxy binary exists
	if !fileExists("/usr/bin/dnscrypt-proxy") {
		return fmt.Errorf("dnscrypt-proxy binary not found at /usr/bin/dnscrypt-proxy")
	}
	
	// Verify config file exists
	if !fileExists("/etc/dnscrypt-proxy/dnscrypt-proxy.toml") {
		return fmt.Errorf("dnscrypt-proxy config not found at /etc/dnscrypt-proxy/dnscrypt-proxy.toml")
	}
	
	cmd := exec.Command("/usr/bin/dnscrypt-proxy", "-config", "/etc/dnscrypt-proxy/dnscrypt-proxy.toml")
	
	// Redirect stdout and stderr to main process stdout for logging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	
	// Configure to run in background
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not start dnscrypt-proxy: %v", err)
	}
	
	log.Printf("dnscrypt-proxy started with PID: %d", cmd.Process.Pid)
	
	// Wait a moment and check if the process is still running
	time.Sleep(1 * time.Second)
	
	// Check if process is still alive
	if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
		log.Printf("Warning: dnscrypt-proxy process may have died: %v", err)
		// Try to get exit status
		if processState, err := cmd.Process.Wait(); err == nil {
			log.Printf("dnscrypt-proxy exit status: %v", processState.String())
		}
		return fmt.Errorf("dnscrypt-proxy process died shortly after startup")
	}
	
	log.Println("dnscrypt-proxy is running successfully")
	return nil
}

func startDnsmasq() error {
	log.Println("Starting dnsmasq in foreground...")
	
	// Verify dnsmasq binary exists
	if !fileExists("/usr/sbin/dnsmasq") {
		return fmt.Errorf("dnsmasq binary not found at /usr/sbin/dnsmasq")
	}
	
	// Verify config file exists
	if !fileExists("/etc/dnsmasq.conf") {
		return fmt.Errorf("dnsmasq config not found at /etc/dnsmasq.conf")
	}
	
	cmd := exec.Command("/usr/sbin/dnsmasq", "--conf-file=/etc/dnsmasq.conf") 
	
	// Redirect stdout and stderr to main process stdout for logging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	
	// Configure to run in background
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	
	// Start dnsmasq and wait for it to complete
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not start dnsmasq: %v", err)
	}
	
	log.Printf("dnsmasq started with PID: %d", cmd.Process.Pid)
	
	// Wait for dnsmasq to complete (it should run indefinitely)
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("dnsmasq exited with error: %v", err)
	}
	
	return nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	
	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func fixFilePermissions() error {
	log.Println("Fixing file permissions for services...")
	
	// Fix permissions for dnscrypt-proxy config file (readable by nobody user)
	log.Printf("Setting permissions for %s", dnscryptConfigPath)
	if err := os.Chmod(dnscryptConfigPath, 0644); err != nil {
		log.Printf("Error fixing permissions for dnscrypt-proxy.toml: %v", err)
		return fmt.Errorf("could not fix permissions for dnscrypt-proxy.toml: %v", err)
	}
	
	// Verify the permissions were actually set
	if info, err := os.Stat(dnscryptConfigPath); err == nil {
		log.Printf("dnscrypt-proxy.toml permissions: %v", info.Mode().Perm())
	}
	
	// Fix permissions for dnsmasq config file (readable by nobody user) 
	log.Printf("Setting permissions for %s", dnsmasqConfPath)
	if err := os.Chmod(dnsmasqConfPath, 0644); err != nil {
		log.Printf("Error fixing permissions for dnsmasq.conf: %v", err)
		return fmt.Errorf("could not fix permissions for dnsmasq.conf: %v", err)
	}
	
	// Verify the permissions were actually set
	if info, err := os.Stat(dnsmasqConfPath); err == nil {
		log.Printf("dnsmasq.conf permissions: %v", info.Mode().Perm())
	}
	
	// Ensure dnscrypt-proxy directories have proper permissions
	configDir := "/etc/dnscrypt-proxy"
	if err := os.Chmod(configDir, 0755); err != nil {
		log.Printf("Warning: could not fix permissions for %s: %v", configDir, err)
	}
	
	// Create and fix permissions for cache directory
	cacheDir := "/var/cache/dnscrypt-proxy"
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Printf("Warning: could not create cache directory %s: %v", cacheDir, err)
	} else {
		// Change ownership to nobody:nobody for cache directory
		if err := os.Chown(cacheDir, 65534, 65534); err != nil {
			log.Printf("Warning: could not change ownership of cache directory: %v", err)
		}
	}
	
	// Create and fix permissions for log directory
	logDir := "/var/log"
	if err := os.Chmod(logDir, 0755); err != nil {
		log.Printf("Warning: could not fix permissions for log directory: %v", err)
	}
	
	log.Println("File permissions fixed successfully")
	return nil
}
