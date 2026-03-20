//go:build linux
// +build linux

package memory

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func sysTotalMemory() uint64 {
	in := &syscall.Sysinfo_t{}
	err := syscall.Sysinfo(in)
	if err != nil {
		return 0
	}
	// If this is a 32-bit system, then these fields are
	// uint32 instead of uint64.
	// So we always convert to uint64 to match signature.
	return uint64(in.Totalram) * uint64(in.Unit)
}

func sysFreeMemory() uint64 {
	in := &syscall.Sysinfo_t{}
	err := syscall.Sysinfo(in)
	if err != nil {
		return 0
	}
	// If this is a 32-bit system, then these fields are
	// uint32 instead of uint64.
	// So we always convert to uint64 to match signature.
	return uint64(in.Freeram) * uint64(in.Unit)
}

func sysAvailableMemory() uint64 {
	fileData, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		fmt.Printf("[Memory]Available memory defaulted to free memory as /proc/meminfo could not be read: %s.\n", err.Error())
		return sysFreeMemory()
	}
	stringData := string(fileData)
	currIndex, endIndex, found := 0, 0, false
	var subSlice string
	for !found && currIndex < len(stringData)-4 {
		subSlice = stringData[currIndex : currIndex+4] //+4: "MemA"
		if subSlice == "MemA" {
			found, endIndex = true, currIndex+len("MemAvailable:")
			for stringData[endIndex] != '\n' {
				endIndex++
			}
		} else {
			for stringData[currIndex] != '\n' {
				currIndex++
			}
			currIndex++
		}
	}
	if found {
		split := strings.Fields(stringData[currIndex:endIndex])
		if len(split) < 2 {
			fmt.Printf("[Memory]Available memory defaulted to free memory as /proc/meminfo MemAvailable line could not be parsed.\n")
			return sysFreeMemory()
		}
		availableKB, _ := strconv.ParseUint(split[1], 10, 64)
		unit := byte(split[2][0])
		switch unit {
		case 'k', 'K':
			return availableKB * 1024
		case 'm', 'M':
			return availableKB * 1024 * 1024
		case 'g', 'G':
			return availableKB * 1024 * 1024 * 1024
		default:
			return availableKB //Bytes
		}
	}
	fmt.Printf("[Memory]Available memory defaulted to free memory as /proc/meminfo MemAvailable line could not be found.\n")
	return sysFreeMemory()
}
