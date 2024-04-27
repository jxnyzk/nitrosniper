package auth


import (
	"fmt"
	"runtime"
	"strings"
	"os/exec"
	"crypto/sha256"
	"encoding/hex"
	"os"
)

func GetCpuID() (string, error) {
	var cpuid string

	switch strings.ToLower(runtime.GOOS) {
	case "linux":
		cpuid = getLinuxSystemID()
	case "windows":
		cpuid = getWindowsCpuID()
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	return cpuid, nil
}

func getLinuxSystemID() string {
    hostname, err := os.Hostname()
    if err != nil {
        return ""
    }

    cpuInfo, err := os.ReadFile("/proc/cpuinfo")
    if err != nil {
        return ""
    }

    cpuInfoStr := string(cpuInfo)
    startIdx := strings.Index(cpuInfoStr, "model name")
    if startIdx == -1 {
        return ""
    }
    
    cpuInfoStr = cpuInfoStr[startIdx:]
    colonIdx := strings.Index(cpuInfoStr, ":")
    if colonIdx == -1 {
        return ""
    }
    
    newlineIdx := strings.Index(cpuInfoStr[colonIdx:], "\n")
    if newlineIdx == -1 {
        return ""
    }
    
    cpuInfoStr = cpuInfoStr[colonIdx+1 : colonIdx+newlineIdx]
    cpuInfoStr = strings.TrimSpace(cpuInfoStr)

    dataToHash := fmt.Sprintf("%s|%s", hostname, cpuInfoStr)
    hasher := sha256.New()
    hasher.Write([]byte(dataToHash))
    systemID := hex.EncodeToString(hasher.Sum(nil))
    
    return systemID
}

func getWindowsCpuID() string {
	cmd := exec.Command("cmd", "/C", "wmic cpu get ProcessorId | findstr /v ProcessorId")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error executing command:", err)
		return ""
	}
	return strings.TrimSpace(string(output))
}