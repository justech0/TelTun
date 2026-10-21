# Builds TelTun.exe: phone helpers first (embedded), then the Wails app.
# Needs Go 1.25+, Node.js and the Wails CLI (go install github.com/wailsapp/wails/v2/cmd/wails@latest).
$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

# Google platform-tools (adb) are bundled into the exe.
if (-not (Test-Path assets/adb/adb.exe)) {
  $zip = Join-Path $env:TEMP "platform-tools.zip"
  Invoke-WebRequest "https://dl.google.com/android/repository/platform-tools-latest-windows.zip" -OutFile $zip
  $tmp = Join-Path $env:TEMP "teltun-pt"
  Expand-Archive $zip $tmp -Force
  foreach ($f in @("adb.exe", "AdbWinApi.dll", "AdbWinUsbApi.dll", "NOTICE.txt")) {
    Copy-Item (Join-Path $tmp "platform-tools/$f") assets/adb/
  }
  Remove-Item $zip, $tmp -Recurse -Force
}

$env:CGO_ENABLED = '0'
$env:GOOS = 'linux'
foreach ($a in @('arm64', 'arm', 'amd64')) {
  $env:GOARCH = $a
  $env:GOARM = '7'
  go build -trimpath -ldflags '-s -w' -o "assets/phone/teltun-$a" ./cmd/teltun-phone
  if ($LASTEXITCODE) { exit 1 }
}
Remove-Item Env:GOOS, Env:GOARCH, Env:GOARM, Env:CGO_ENABLED

# sing-box needs these tags and -checklinkname=0 (see its release/LDFLAGS).
# Wails' binding step ignores -ldflags, so bindings are kept in the repo.
$tags = 'with_gvisor badlinkname tfogo_checklinkname0'
wails build -clean -skipbindings -trimpath -tags $tags `
  -ldflags '-s -w -checklinkname=0 -X internal/godebug.defaultGODEBUG=multipathtcp=0'
if ($LASTEXITCODE) { exit 1 }
Write-Host "Done: build\bin\TelTun.exe"
