package deploy

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"v1-sg-deployment-tool/internal/models"
)

type InstallMethod string

const (
	MethodCurlDownload       InstallMethod = "curl_download"
	MethodPowerShellDownload InstallMethod = "powershell_download"
)

type PackageType string

const (
	PackageTypeBinary PackageType = "binary"
	PackageTypeMSI    PackageType = "msi"
	PackageTypeEXE    PackageType = "exe"
	PackageTypePKG    PackageType = "pkg"
	PackageTypeDEB    PackageType = "deb"
	PackageTypeRPM    PackageType = "rpm"
)

type InstallRequest struct {
	OS               models.TargetOS
	BinaryURL        string
	DestinationPath  string
	PostInstallArgs  []string
	ExecuteOnInstall bool
	PackageType      PackageType
	Checksum         string
	ChecksumAlg      string
	ExpectedArch     string
	MinFreeMB        int
	ProxyURL         string
	RequiresReboot   bool
	AllowReboot      bool
}

type DeployPlan struct {
	Method   InstallMethod
	Commands []string
}

func BuildPlan(request InstallRequest) (DeployPlan, error) {
	if request.BinaryURL == "" {
		return DeployPlan{}, errors.New("binary url is required")
	}

	if err := validateInstallerURL(request.BinaryURL); err != nil {
		return DeployPlan{}, err
	}

	if request.DestinationPath == "" {
		request.DestinationPath = defaultDestinationPath(request.OS)
	}

	if request.PackageType == "" {
		request.PackageType = packageTypeFromURL(request.BinaryURL)
	}

	if !isPackageTypeAllowed(request.OS, request.PackageType) {
		return DeployPlan{}, errors.New("package type does not match target os")
	}

	switch request.OS {
	case models.TargetOSLinux, models.TargetOSMacOS:
		return buildUnixPlan(request), nil
	case models.TargetOSWindows:
		return buildWindowsPlan(request), nil
	default:
		return DeployPlan{}, errors.New("unsupported os")
	}
}

func validateInstallerURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return errors.New("binary url is invalid")
	}
	if parsed.Hostname() == "" {
		return errors.New("binary url must include a host")
	}
	return nil
}

func buildUnixPlan(request InstallRequest) DeployPlan {
	folderPath, filePath := splitPath(request.DestinationPath)
	commands := []string{"set -e"}
	if request.MinFreeMB > 0 {
		commands = append(commands, unixDiskCheckCommand(folderPath, request.MinFreeMB))
	}
	if request.ExpectedArch != "" {
		commands = append(commands, unixArchCheckCommand(request.ExpectedArch))
	}

	download := unixDownloadCommand(request.BinaryURL, filePath, request.ProxyURL)
	commands = append(commands, "mkdir -p \""+folderPath+"\"", download, "chmod +x \""+filePath+"\"")

	if request.Checksum != "" {
		commands = append(commands, unixChecksumCommand(request.OS, filePath, request.Checksum))
	}

	installCommand, supportsInstall := unixInstallCommand(request, filePath)
	if supportsInstall {
		commands = append(commands, installCommand)
	}

	if request.ExecuteOnInstall {
		commands = append(commands, buildRunCommand(filePath, request.PostInstallArgs))
	}

	if request.ExecuteOnInstall || supportsInstall {
		if request.RequiresReboot && request.AllowReboot {
			commands = append(commands, "sudo reboot")
		}
		commands = append(commands, "rm -f \""+filePath+"\"", "rmdir \""+folderPath+"\" 2>/dev/null || true")
	}

	return DeployPlan{
		Method:   MethodCurlDownload,
		Commands: commands,
	}
}

func buildWindowsPlan(request InstallRequest) DeployPlan {
	folderPath, filePath := splitPath(request.DestinationPath)
	createFolder := "powershell -NoProfile -Command \"New-Item -ItemType Directory -Force -Path '" + folderPath + "'\""
	diskCheck := windowsDiskCheckCommand(folderPath, request.MinFreeMB)
	archCheck := windowsArchCheckCommand(request.ExpectedArch)
	download := windowsDownloadCommand(request.BinaryURL, filePath, request.ProxyURL)
	unblock := "powershell -NoProfile -Command \"Unblock-File -Path '" + filePath + "'\""
	removeFile := "powershell -NoProfile -Command \"Remove-Item -Path '" + filePath + "' -Force\""
	removeFolder := "powershell -NoProfile -Command \"Remove-Item -Path '" + folderPath + "' -Force\""

	commands := []string{}
	if diskCheck != "" {
		commands = append(commands, diskCheck)
	}
	if archCheck != "" {
		commands = append(commands, archCheck)
	}
	commands = append(commands, createFolder, download, unblock)

	if request.Checksum != "" {
		commands = append(commands, windowsChecksumCommand(filePath, request.Checksum))
	}

	installCommand, supportsInstall := windowsInstallCommand(request, filePath)
	if supportsInstall {
		commands = append(commands, installCommand)
	}
	if request.ExecuteOnInstall {
		commands = append(commands, buildRunCommand(filePath, request.PostInstallArgs))
	}
	if request.ExecuteOnInstall || supportsInstall {
		if request.RequiresReboot && request.AllowReboot {
			commands = append(commands, "powershell -NoProfile -Command \"Restart-Computer -Force\"")
		}
		commands = append(commands, removeFile, removeFolder)
	}

	return DeployPlan{
		Method:   MethodPowerShellDownload,
		Commands: commands,
	}
}

func buildRunCommand(path string, args []string) string {
	if len(args) == 0 {
		return "\"" + path + "\""
	}

	escapedArgs := make([]string, 0, len(args))
	for _, arg := range args {
		escapedArgs = append(escapedArgs, "\""+arg+"\"")
	}

	return "\"" + path + "\" " + strings.Join(escapedArgs, " ")
}

func defaultDestinationPath(os models.TargetOS) string {
	switch os {
	case models.TargetOSWindows:
		return `C:\V1SGDeploymentTool\installer.bin`
	case models.TargetOSLinux, models.TargetOSMacOS:
		return "/tmp/V1SGDeploymentTool/installer.bin"
	default:
		return "/tmp/V1SGDeploymentTool/installer.bin"
	}
}

func splitPath(path string) (string, string) {
	lastSlash := strings.LastIndex(path, "/")
	lastBackslash := strings.LastIndex(path, "\\")
	lastIndex := lastSlash
	if lastBackslash > lastIndex {
		lastIndex = lastBackslash
	}
	if lastIndex == -1 {
		return ".", path
	}

	return path[:lastIndex], path
}

func unixDownloadCommand(url string, path string, proxyURL string) string {
	if proxyURL == "" {
		return "curl -fsSL \"" + url + "\" -o \"" + path + "\""
	}
	return "HTTPS_PROXY=\"" + proxyURL + "\" HTTP_PROXY=\"" + proxyURL + "\" curl -fsSL \"" + url + "\" -o \"" + path + "\""
}

func windowsDownloadCommand(url string, path string, proxyURL string) string {
	if proxyURL == "" {
		return "powershell -NoProfile -Command \"Invoke-WebRequest -Uri '" + url + "' -OutFile '" + path + "'\""
	}
	return "powershell -NoProfile -Command \"$env:HTTPS_PROXY='" + proxyURL + "'; $env:HTTP_PROXY='" + proxyURL + "'; Invoke-WebRequest -Uri '" + url + "' -OutFile '" + path + "'\""
}

func unixChecksumCommand(os models.TargetOS, path string, checksum string) string {
	if os == models.TargetOSMacOS {
		return "echo \"" + checksum + "  " + path + "\" | shasum -a 256 -c -"
	}
	return "echo \"" + checksum + "  " + path + "\" | sha256sum -c -"
}

func windowsChecksumCommand(path string, checksum string) string {
	return "powershell -NoProfile -Command \"$hash=(Get-FileHash -Algorithm SHA256 -Path '" + path + "').Hash.ToLower(); if ($hash -ne '" + strings.ToLower(checksum) + "') { throw 'checksum_mismatch' }\""
}

func unixArchCheckCommand(expectedArch string) string {
	return "arch=$(uname -m); if [ \"$arch\" != \"" + expectedArch + "\" ] && [ \"$arch\" != \"" + normalizeArch(expectedArch) + "\" ]; then echo \"arch_mismatch\"; exit 1; fi"
}

func windowsArchCheckCommand(expectedArch string) string {
	if expectedArch == "" {
		return ""
	}
	return "powershell -NoProfile -Command \"$arch=$env:PROCESSOR_ARCHITECTURE; if ($arch -ne '" + normalizeArch(expectedArch) + "') { throw 'arch_mismatch' }\""
}

func unixDiskCheckCommand(folderPath string, minFreeMB int) string {
	return "avail=$(df -Pm \"" + folderPath + "\" | awk 'NR==2 {print $4}'); if [ \"$avail\" -lt \"" + fmt.Sprintf("%d", minFreeMB) + "\" ]; then echo \"insufficient_disk\"; exit 1; fi"
}

func windowsDiskCheckCommand(folderPath string, minFreeMB int) string {
	if minFreeMB <= 0 {
		return ""
	}
	drive := "C"
	if len(folderPath) >= 2 && folderPath[1] == ':' {
		drive = folderPath[:1]
	}
	return "powershell -NoProfile -Command \"$free=((Get-PSDrive -Name '" + drive + "').Free/1MB); if ($free -lt " + fmt.Sprintf("%d", minFreeMB) + ") { throw 'insufficient_disk' }\""
}

func unixInstallCommand(request InstallRequest, path string) (string, bool) {
	switch request.PackageType {
	case PackageTypeDEB:
		return "sudo dpkg -i \"" + path + "\"", true
	case PackageTypeRPM:
		return "sudo rpm -Uvh \"" + path + "\"", true
	case PackageTypePKG:
		return "sudo installer -pkg \"" + path + "\" -target /", true
	case PackageTypeBinary:
		return "", false
	default:
		return "", false
	}
}

func windowsInstallCommand(request InstallRequest, path string) (string, bool) {
	switch request.PackageType {
	case PackageTypeMSI:
		return "powershell -NoProfile -Command \"Start-Process msiexec -ArgumentList '/i','" + path + "','/qn','/norestart' -Wait\"", true
	case PackageTypeEXE:
		return "powershell -NoProfile -Command \"Start-Process -FilePath '" + path + "' -ArgumentList '/quiet','/norestart' -Wait\"", true
	case PackageTypeBinary:
		return "", false
	default:
		return "", false
	}
}

func normalizeArch(value string) string {
	switch strings.ToLower(value) {
	case "x86_64":
		return "amd64"
	case "amd64":
		return "amd64"
	case "arm64":
		return "arm64"
	default:
		return strings.ToLower(value)
	}
}

func isPackageTypeAllowed(os models.TargetOS, packageType PackageType) bool {
	if packageType == "" || packageType == PackageTypeBinary {
		return true
	}

	switch os {
	case models.TargetOSWindows:
		return packageType == PackageTypeMSI || packageType == PackageTypeEXE
	case models.TargetOSMacOS:
		return packageType == PackageTypePKG
	case models.TargetOSLinux:
		return packageType == PackageTypeDEB || packageType == PackageTypeRPM
	default:
		return false
	}
}

func packageTypeFromURL(url string) PackageType {
	lower := strings.ToLower(url)
	switch {
	case strings.HasSuffix(lower, ".msi"):
		return PackageTypeMSI
	case strings.HasSuffix(lower, ".exe"):
		return PackageTypeEXE
	case strings.HasSuffix(lower, ".pkg"):
		return PackageTypePKG
	case strings.HasSuffix(lower, ".deb"):
		return PackageTypeDEB
	case strings.HasSuffix(lower, ".rpm"):
		return PackageTypeRPM
	default:
		return PackageTypeBinary
	}
}
