[CmdletBinding()]
param(
    [Parameter(Position = 0)]
    [ValidateSet('help', 'build', 'run', 'test', 'clean', 'docker-build', 'docker-run', 'docker-test')]
    [string]$Command = 'help'
)

$ErrorActionPreference = 'Stop'
# Handle native exit codes explicitly, including in PowerShell 7.
$PSNativeCommandUseErrorActionPreference = $false

function Invoke-CheckedCommand {
    param(
        [string]$Program,
        [string[]]$Arguments
    )

    & $Program @Arguments
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}

function Invoke-DockerStack {
    param([string[]]$UpArguments)

    $exitCode = 0
    try {
        & docker compose up @UpArguments
        $exitCode = $LASTEXITCODE
    }
    finally {
        # PowerShell skips the rest of the try block on Ctrl+C.
        # Keep report copying in finally so interrupted runs preserve reports too.
        try {
            New-Item -ItemType Directory -Path 'reports' -Force | Out-Null
            # Copy through the Docker API so remote daemons (including DinD) work too.
            & docker compose cp 'perf-test:/app/reports/.' 'reports/'
            if ($LASTEXITCODE -ne 0) {
                if ($exitCode -eq 0) { $exitCode = 1 }
                [Console]::Error.WriteLine('Could not copy reports from the perf-test container.')
            }
            else {
                Write-Host 'Reports saved to reports/'
            }
        }
        catch {
            [Console]::Error.WriteLine($_.Exception.Message)
            if ($exitCode -eq 0) { $exitCode = 1 }
        }
    }
    exit $exitCode
}

if ($Command -eq 'help') {
    @'
Usage: .\perf.ps1 [command]

  help          Show available commands (default)
  build         Build the framework locally
  run           Start the mock server on port 8080
  test          Run local performance tests using config.yaml
  clean         Remove build artifacts and reports
  docker-build  Build Docker images
  docker-run    Start the full stack in Docker
  docker-test   Run performance tests in Docker
'@
    exit 0
}

$dependency = switch ($Command) {
    { $_ -in 'build', 'run', 'test' } { 'go' }
    { $_ -like 'docker-*' } { 'docker' }
}
if ($dependency -and -not (Get-Command $dependency -CommandType Application -ErrorAction SilentlyContinue)) {
    [Console]::Error.WriteLine("Required command not found: $dependency")
    exit 127
}

Push-Location -LiteralPath $PSScriptRoot
try {
    switch ($Command) {
        'build' {
            Write-Host 'Building framework...'
            $targetOS = & go env GOOS
            if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
            $binary = if ($targetOS -eq 'windows') { 'bin/perf-runner.exe' } else { 'bin/perf-runner' }
            New-Item -ItemType Directory -Path 'bin' -Force | Out-Null
            Invoke-CheckedCommand -Program 'go' -Arguments @('build', '-o', $binary, './cmd/runner')
            Write-Host "Build complete: $binary"
        }
        'run' {
            Write-Host 'Starting mock server on port 8080...'
            Write-Host 'Press Ctrl+C to stop'
            Set-Location -LiteralPath (Join-Path $PSScriptRoot 'mock-server')
            Invoke-CheckedCommand -Program 'go' -Arguments @('run', 'main.go')
        }
        'test' {
            Write-Host 'Running performance tests...'
            Invoke-CheckedCommand -Program 'go' -Arguments @('run', './cmd/runner', '-config', 'config.yaml')
        }
        'clean' {
            Write-Host 'Cleaning artifacts...'
            foreach ($path in @('bin', 'reports')) {
                if (Test-Path -LiteralPath $path) {
                    Remove-Item -LiteralPath $path -Recurse -Force
                }
            }
            Write-Host 'Clean complete.'
        }
        'docker-build' {
            Write-Host 'Building Docker images...'
            Invoke-CheckedCommand -Program 'docker' -Arguments @('compose', 'build')
        }
        'docker-run' {
            Write-Host 'Starting full stack in Docker...'
            Invoke-DockerStack -UpArguments @('--build')
        }
        'docker-test' {
            Write-Host 'Running tests in Docker...'
            Invoke-DockerStack -UpArguments @('--build', '--force-recreate', '--abort-on-container-exit', '--exit-code-from', 'perf-test')
        }
    }
}
finally {
    Pop-Location
}
exit 0
