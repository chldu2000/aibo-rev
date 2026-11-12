$ErrorActionPreference = 'Stop'
$probeRoot = Split-Path $PSScriptRoot -Parent
$probeExe = Join-Path $probeRoot 'bin/validate.exe'
$results = @()
foreach ($mode in @('force', 'taskkill', 'normal')) {
    $reportPath = Join-Path $probeRoot "reports/win-children-$mode.json"
    if (Test-Path -LiteralPath $reportPath) { Remove-Item -LiteralPath $reportPath }
    $probe = Start-Process -FilePath $probeExe -ArgumentList @('-children', "`"$reportPath`"") -WindowStyle Hidden -PassThru
    $deadline = (Get-Date).AddSeconds(15)
    while (!(Test-Path -LiteralPath $reportPath)) {
        if ($probe.HasExited) { throw "Probe exited before reporting: $($probe.ExitCode)" }
        if ((Get-Date) -gt $deadline) { throw 'Probe report timed out' }
        Start-Sleep -Milliseconds 100
    }
    $report = Get-Content -Raw -LiteralPath $reportPath | ConvertFrom-Json
    if ($report.app -ne $probe.Id) { throw 'Unexpected probe identity' }
    foreach ($child in $report.children) {
        if (!(Get-Process -Id $child.pid -ErrorAction SilentlyContinue)) { throw "Child not alive before experiment: $($child.pid)" }
    }
    if ($mode -eq 'force') { Stop-Process -Id $probe.Id -Force }
    if ($mode -eq 'taskkill') {
        taskkill.exe /F /PID $probe.Id
        if ($LASTEXITCODE -ne 0) { throw 'taskkill failed' }
    }
    if (!$probe.WaitForExit(15000)) { throw 'Probe did not exit' }
    $deadline = (Get-Date).AddSeconds(5)
    do {
        $remaining = @($report.children | Where-Object { Get-Process -Id $_.pid -ErrorAction SilentlyContinue })
        if ($remaining.Count -eq 0) { break }
        Start-Sleep -Milliseconds 100
    } while ((Get-Date) -lt $deadline)
    $results += [pscustomobject]@{ mode=$mode; app=$report.app; children=$report.children; remaining=$remaining; passed=($remaining.Count -eq 0) }
}
$results | ConvertTo-Json -Depth 8 | Set-Content -Encoding UTF8 -LiteralPath (Join-Path $probeRoot 'reports/win-children-result.json')
if (@($results | Where-Object { !$_.passed }).Count) { throw 'Child processes survived' }
$results | ConvertTo-Json -Depth 8
