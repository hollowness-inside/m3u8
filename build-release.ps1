# Parameters
$AppName = "m3u8"
$Version = "v1.0.0"
$OutputDir = "release"
$MainPath = ".\cmd\m3u8"

# Target platforms
$Platforms = @(
    "linux/amd64",
    "linux/arm64",
    "windows/amd64",
    "darwin/amd64",
    "darwin/arm64"
)

# Create release directory
New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

foreach ($Platform in $Platforms) {
    $Parts = $Platform -split '/'
    $GOOS = $Parts[0]
    $GOARCH = $Parts[1]

    # Set output file name
    $OutputName = "$AppName-$Version-$GOOS-$GOARCH"
    if ($GOOS -eq "windows") {
        $OutputName += ".exe"
    }

    New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

    # Build the binary
    Write-Host "Building for $GOOS/$GOARCH..."
    $env:GOOS = $GOOS
    $env:GOARCH = $GOARCH
    $CurrentDir = (Get-Location).Path
    & go build -C $CurrentDir -o (Join-Path $OutputDir $OutputName) -ldflags "-s -X main.version=$Version" $MainPath

    if ($LASTEXITCODE -ne 0) {
        Write-Error "Build failed for $GOOS/$GOARCH"
        continue
    }
}

Write-Host "Build complete. Files are in the '$OutputDir' directory."
