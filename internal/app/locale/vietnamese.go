package locale

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// ContainsVietnamese checks if the input contains Vietnamese characters
func ContainsVietnamese(s string) bool {
	for _, r := range s {
		if isVietnameseChar(r) {
			return true
		}
	}
	return false
}

// isVietnameseChar checks if a rune is a Vietnamese-specific character
func isVietnameseChar(r rune) bool {
	// Vietnamese uses Latin alphabet with additional diacritics
	// Check for Vietnamese-specific characters (beyond basic ASCII)
	vietnameseChars := "àáảãạăằắẳẵặâầấẩẫậèéẻẽẹêềếểễệìíỉĩịòóỏõọôồốổỗộơờớởỡợùúủũụưừứửữựỳýỷỹỵđ"
	vietnameseChars += strings.ToUpper(vietnameseChars)

	return strings.ContainsRune(vietnameseChars, r)
}

// CheckVietnameseFontSupport checks if the system likely supports Vietnamese fonts
// Returns (supported, suggestion)
func CheckVietnameseFontSupport() (bool, string) {
	switch runtime.GOOS {
	case "linux":
		return checkLinuxFontSupport()
	case "darwin":
		// macOS typically has good Unicode support
		return true, ""
	case "windows":
		// Windows typically has good Unicode support with modern terminals
		return true, ""
	default:
		return true, ""
	}
}

func checkLinuxFontSupport() (bool, string) {
	// Check locale
	locale := os.Getenv("LANG")
	if locale == "" {
		locale = os.Getenv("LC_ALL")
	}

	// Check if locale is UTF-8
	isUTF8 := strings.Contains(strings.ToLower(locale), "utf-8") ||
		strings.Contains(strings.ToLower(locale), "utf8")

	if !isUTF8 {
		return false, `Hệ thống chưa cấu hình UTF-8. Để hỗ trợ tiếng Việt, chạy:
  sudo apt-get install locales
  sudo locale-gen vi_VN.UTF-8
  export LANG=vi_VN.UTF-8`
	}

	// Check if Vietnamese locale is available
	output, err := exec.Command("locale", "-a").Output()
	if err == nil {
		locales := string(output)
		hasVietnamese := strings.Contains(locales, "vi_VN") ||
			strings.Contains(locales, "vietnamese")

		if !hasVietnamese {
			return false, `Hệ thống chưa có locale tiếng Việt. Để cài đặt:
  sudo apt-get install locales
  sudo locale-gen vi_VN.UTF-8`
		}
	}

	// Check if common Unicode fonts are installed
	fontDirs := []string{
		"/usr/share/fonts/truetype/dejavu",
		"/usr/share/fonts/truetype/noto",
		"/usr/share/fonts/truetype/liberation",
	}

	hasUnicodeFont := false
	for _, dir := range fontDirs {
		if _, err := os.Stat(dir); err == nil {
			hasUnicodeFont = true
			break
		}
	}

	if !hasUnicodeFont {
		return false, `Có thể thiếu font Unicode. Để cài đặt font hỗ trợ tiếng Việt:
  sudo apt-get install fonts-dejavu fonts-noto
  # hoặc
  sudo apt-get install fonts-liberation`
	}

	return true, ""
}

// WarnIfVietnameseNotSupported prints a warning if Vietnamese may not display correctly
func WarnIfVietnameseNotSupported(input string) {
	if !ContainsVietnamese(input) {
		return
	}

	supported, suggestion := CheckVietnameseFontSupport()
	if !supported && suggestion != "" {
		// Use ASCII-safe warning
		println("\n[!] Canh bao: He thong co the khong hien thi tieng Viet dung.")
		println(suggestion)
		println("")
	}
}
