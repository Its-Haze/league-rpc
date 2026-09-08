# Builds the .syso carrying the icon, manifest and version resources.
# A release passes -Version so the tag, not config.yml, stamps the binary.
param(
    [Parameter(Mandatory = $true)][string]$Arch,
    [Parameter(Mandatory = $true)][string]$Out,
    [string]$Version
)

$ErrorActionPreference = 'Stop'
$windowsDir = $PSScriptRoot
$infoPath = Join-Path $windowsDir 'info.json'

$version = $Version.Trim() -replace '^v', '' -replace '-.*$', ''
if ($version) {
    $info = Get-Content $infoPath -Raw | ConvertFrom-Json
    $info.fixed.file_version = $version
    $info.fixed.product_version = $version
    $info.info.'0409'.ProductVersion = $version

    $infoPath = Join-Path $env:TEMP ('league-rpc-info-' + [guid]::NewGuid() + '.json')
    [IO.File]::WriteAllText($infoPath, ($info | ConvertTo-Json -Depth 10), (New-Object Text.UTF8Encoding $false))
}

try {
    wails3 generate syso -arch $Arch `
        -icon (Join-Path $windowsDir 'icon.ico') `
        -manifest (Join-Path $windowsDir 'wails.exe.manifest') `
        -info $infoPath `
        -out $Out
    if ($LASTEXITCODE -ne 0) { throw "wails3 generate syso failed with exit code $LASTEXITCODE" }
}
finally {
    if ($version) { Remove-Item $infoPath -Force -ErrorAction SilentlyContinue }
}
