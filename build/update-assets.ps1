# Regenerates the Windows build assets from config.yml. wails3 also emits
# darwin/linux/ios output, so it runs in a scratch dir and only Windows returns.
param([Parameter(Mandatory = $true)][string]$AppName)

$ErrorActionPreference = 'Stop'
$buildDir = $PSScriptRoot
$tmp = Join-Path $env:TEMP ('wails-assets-' + [guid]::NewGuid())

New-Item -ItemType Directory -Path $tmp | Out-Null
try {
    Copy-Item (Join-Path $buildDir 'config.yml') $tmp
    Push-Location $tmp
    try {
        wails3 update build-assets -name $AppName -binaryname $AppName -config config.yml -dir .
        if ($LASTEXITCODE -ne 0) { throw "wails3 update build-assets failed with exit code $LASTEXITCODE" }
    }
    finally { Pop-Location }

    $generated = @(
        'windows/info.json'
        'windows/nsis/wails_tools.nsh'
        'windows/wails.exe.manifest'
    )
    foreach ($file in $generated) {
        Copy-Item (Join-Path $tmp $file) (Join-Path $buildDir $file) -Force
        Write-Host "updated build/$file"
    }

    # wails3 keys the string table to language 0000, which leaves Explorer's
    # Details tab blank; 0409 (en-US) is what the Windows reader resolves.
    $infoPath = Join-Path $buildDir 'windows/info.json'
    $info = Get-Content $infoPath -Raw | ConvertFrom-Json
    if ($info.info.PSObject.Properties.Name -contains '0000') {
        $strings = $info.info.'0000'
        $info.info.PSObject.Properties.Remove('0000')
        $info.info | Add-Member -NotePropertyName '0409' -NotePropertyValue $strings
    }
    # Without this the Details tab reports a product version of 0.0.0.0.
    if (-not $info.fixed.PSObject.Properties.Name.Contains('product_version')) {
        $info.fixed | Add-Member -NotePropertyName 'product_version' -NotePropertyValue $info.fixed.file_version
    }
    # Set-Content -Encoding UTF8 writes a BOM on Windows PowerShell, which the
    # Go side of wails3 will not parse.
    [IO.File]::WriteAllText($infoPath, ($info | ConvertTo-Json -Depth 10), (New-Object Text.UTF8Encoding $false))
    Write-Host "retargeted build/windows/info.json to language 0409"
}
finally { Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue }
