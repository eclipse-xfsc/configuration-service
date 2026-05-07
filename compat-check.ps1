Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

Push-Location $PSScriptRoot
try {
	go -C ./.ai run . --repo-root .. validate
	go test ./...
}
finally {
	Pop-Location
}
