package priority

// PriorityEntry represents a configured process priority
type PriorityEntry struct {
	ProcessName      string // e.g., "LeagueClient.exe"
	CpuPriority      int    // 1-6 (no 4)
	IoPriority       int    // 0-3
	PagePriority     int    // 0-5
	CpuPriorityName  string // "High", "Above Normal", etc.
	IoPriorityName   string
	PagePriorityName string
}

// GetCpuPriorityName returns human-readable name for CPU priority value
func GetCpuPriorityName(value int) string {
	switch value {
	case 1:
		return "Idle"
	case 2:
		return "Normal"
	case 3:
		return "High"
	case 5:
		return "Below Normal"
	case 6:
		return "Above Normal"
	default:
		return "Unknown"
	}
}

// GetIoPriorityName returns human-readable name for I/O priority value
func GetIoPriorityName(value int) string {
	switch value {
	case 0:
		return "Very Low"
	case 1:
		return "Low"
	case 2:
		return "Normal"
	case 3:
		return "High"
	default:
		return "Unknown"
	}
}

// GetPagePriorityName returns human-readable name for page priority value
func GetPagePriorityName(value int) string {
	switch value {
	case 0:
		return "Idle"
	case 1:
		return "Very Low"
	case 2:
		return "Low"
	case 3:
		return "Background"
	case 4:
		return "Default"
	case 5:
		return "Normal"
	default:
		return "Unknown"
	}
}

// ParseCpuPriorityName converts name to priority value
func ParseCpuPriorityName(name string) int {
	switch name {
	case "idle", "Idle":
		return 1
	case "normal", "Normal":
		return 2
	case "high", "High":
		return 3
	case "below-normal", "Below Normal", "below normal":
		return 5
	case "above-normal", "Above Normal", "above normal":
		return 6
	default:
		return 2 // default to normal
	}
}

// ParseIoPriorityName converts name to I/O priority value
func ParseIoPriorityName(name string) int {
	switch name {
	case "very-low", "Very Low", "very low":
		return 0
	case "low", "Low":
		return 1
	case "normal", "Normal":
		return 2
	case "high", "High":
		return 3
	default:
		return 2 // default to normal
	}
}

// ParsePagePriorityName converts name to page priority value
func ParsePagePriorityName(name string) int {
	switch name {
	case "idle", "Idle":
		return 0
	case "very-low", "Very Low", "very low":
		return 1
	case "low", "Low":
		return 2
	case "background", "Background":
		return 3
	case "default", "Default":
		return 4
	case "normal", "Normal":
		return 5
	default:
		return 5 // default to normal
	}
}
