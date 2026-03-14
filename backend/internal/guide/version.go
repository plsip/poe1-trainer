package guide

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const guideVersionEnvVar = "GUIDE_VERSION"

// ResolveVersion returns the explicit version when provided, otherwise tries to
// resolve the current git commit hash or an override from GUIDE_VERSION.
func ResolveVersion(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	if value := strings.TrimSpace(os.Getenv(guideVersionEnvVar)); value != "" {
		return value, nil
	}
	out, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err == nil {
		if value := strings.TrimSpace(string(out)); value != "" {
			return value, nil
		}
	}
	return "", fmt.Errorf("guide version not provided and git commit hash unavailable; pass version explicitly or set %s", guideVersionEnvVar)
}