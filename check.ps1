param(
    [string]$RepoRoot = (Get-Location).Path,
    [string]$OutRoot = ''
)

Set-StrictMode -Version 2.0
$ErrorActionPreference = 'Stop'
$script:CheckTimeoutSec = 124

function Get-CheckGoCommand {
    $preferred = 'k:\go\go1.20.14\bin\go.exe'
    if (Test-Path -LiteralPath $preferred) {
        return (Resolve-Path -LiteralPath $preferred).Path
    }
    $go = Get-Command go -ErrorAction SilentlyContinue
    if ($go) {
        return $go.Source
    }
    throw 'go executable not found'
}

function Get-CheckMakeCommand {
    $make = Get-Command make -ErrorAction SilentlyContinue
    if ($make) {
        return $make.Source
    }
    return ''
}

function Get-CheckGitCommand {
    $git = Get-Command git -ErrorAction SilentlyContinue
    if (-not $git) {
        throw 'git executable not found'
    }
    return $git.Source
}

function New-CheckContext {
    param(
        [Parameter(Mandatory=$true)][string]$Root,
        [string]$OutRoot = ''
    )

    $resolvedRoot = (Resolve-Path -LiteralPath $Root).Path
    if ($OutRoot -eq '') {
        $OutRoot = Join-Path $resolvedRoot '.check-results'
    }

    $stamp = Get-Date -Format 'yyyyMMdd_HHmmss'
    $resultDir = Join-Path $OutRoot "policy_playground_check_$stamp"
    $ctx = [ordered]@{
        RepoRoot = $resolvedRoot
        ResultDir = $resultDir
        LogsDir = Join-Path $resultDir 'logs'
        InputsDir = Join-Path $resultDir 'inputs'
        OutputsDir = Join-Path $resultDir 'outputs'
        MetaDir = Join-Path $resultDir 'meta'
        TmpDir = Join-Path $resultDir 'tmp'
        CommandsPath = Join-Path $resultDir 'commands.jsonl'
        Assessments = New-Object System.Collections.ArrayList
        CommandResults = @{}
        StartedAt = (Get-Date).ToString('o')
        GoCmd = Get-CheckGoCommand
        MakeCmd = Get-CheckMakeCommand
        GitCmd = Get-CheckGitCommand
    }

    foreach ($directory in @($ctx.ResultDir, $ctx.LogsDir, $ctx.InputsDir, $ctx.OutputsDir, $ctx.MetaDir, $ctx.TmpDir)) {
        New-Item -ItemType Directory -Force -Path $directory | Out-Null
    }
    '' | Set-Content -LiteralPath $ctx.CommandsPath -Encoding UTF8
    '' | Set-Content -LiteralPath (Join-Path $ctx.MetaDir 'git_status_short.txt') -Encoding UTF8
    return $ctx
}

function Write-CheckText {
    param(
        [Parameter(Mandatory=$true)]$Ctx,
        [Parameter(Mandatory=$true)][string]$RelativePath,
        [Parameter(Mandatory=$true)][string]$Content
    )

    $path = Join-Path $Ctx.ResultDir $RelativePath
    $parent = Split-Path -Parent $path
    if ($parent) {
        New-Item -ItemType Directory -Force -Path $parent | Out-Null
    }
    Set-Content -LiteralPath $path -Value $Content -Encoding UTF8
    return $path
}

function Save-CheckJson {
    param(
        [Parameter(Mandatory=$true)][string]$Path,
        [Parameter(Mandatory=$true)]$Value
    )

    $json = $Value | ConvertTo-Json -Depth 60
    Set-Content -LiteralPath $Path -Value $json -Encoding UTF8
}

function Get-RelativeResultPath {
    param(
        [Parameter(Mandatory=$true)]$Ctx,
        [Parameter(Mandatory=$true)][string]$Path
    )

    if ([string]::IsNullOrWhiteSpace($Path)) {
        return ''
    }
    $candidate = $Path
    if (-not [System.IO.Path]::IsPathRooted($candidate)) {
        $candidate = Join-Path $Ctx.ResultDir $candidate
    }
    if (-not (Test-Path -LiteralPath $candidate)) {
        return ''
    }
    $full = (Resolve-Path -LiteralPath $candidate).Path
    if ($full.StartsWith($Ctx.ResultDir, [System.StringComparison]::OrdinalIgnoreCase)) {
        return $full.Substring($Ctx.ResultDir.Length).TrimStart('\').Replace('\', '/')
    }
    return $full.Replace('\', '/')
}

function Get-CheckProcessChildren {
    param([int]$ParentId)

    $children = @(Get-CimInstance Win32_Process -Filter "ParentProcessId = $ParentId" -ErrorAction SilentlyContinue)
    $ids = New-Object System.Collections.Generic.List[int]
    foreach ($child in $children) {
        $childId = [int]$child.ProcessId
        $ids.Add($childId)
        $descendants = Get-CheckProcessChildren -ParentId $childId
        foreach ($id in $descendants) {
            $ids.Add([int]$id)
        }
    }
    return @($ids | Sort-Object -Unique)
}

function Get-CheckProcessTreeStats {
    param([int]$RootPid)

    $ids = @(Get-CheckProcessChildren -ParentId $RootPid)
    $ids += $RootPid
    $ids = @($ids | Sort-Object -Unique)
    $workingSet = [int64]0
    foreach ($id in $ids) {
        try {
            $process = Get-Process -Id $id -ErrorAction Stop
            $workingSet += [int64]$process.WorkingSet64
        } catch {
        }
    }
    return [ordered]@{
        ids = $ids
        working_set_bytes = $workingSet
    }
}

function Stop-CheckProcessTree {
    param([int]$RootPid)

    $ids = @(Get-CheckProcessChildren -ParentId $RootPid)
    $ids += $RootPid
    $ids = @($ids | Sort-Object -Descending -Unique)
    foreach ($id in $ids) {
        try {
            Stop-Process -Id $id -Force -ErrorAction SilentlyContinue
        } catch {
        }
    }
}

function Invoke-CheckHiddenCommand {
    param(
        [Parameter(Mandatory=$true)]$Ctx,
        [Parameter(Mandatory=$true)][string]$Name,
        [Parameter(Mandatory=$true)][string]$FilePath,
        [string[]]$Arguments = @(),
        [string]$WorkingDirectory = '',
        [int]$TimeoutSec = $script:CheckTimeoutSec,
        [string]$MonitorPath = '',
        [bool]$ExpectedFailure = $false,
        [string]$ExitCodeSidecarPath = ''
    )

    if ($WorkingDirectory -eq '') {
        $WorkingDirectory = $Ctx.RepoRoot
    }

    $safeName = $Name -replace '[^A-Za-z0-9_.-]', '_'
    $stdoutPath = Join-Path $Ctx.LogsDir "$safeName.stdout.log"
    $stderrPath = Join-Path $Ctx.LogsDir "$safeName.stderr.log"
    $logPath = Join-Path $Ctx.LogsDir "$safeName.log"

    $started = Get-Date
    $process = Start-Process -FilePath $FilePath -ArgumentList $Arguments -WorkingDirectory $WorkingDirectory -WindowStyle Hidden -RedirectStandardOutput $stdoutPath -RedirectStandardError $stderrPath -PassThru

    $peakTreeWorkingSetBytes = [int64]0
    $timedOut = $false
    $samples = New-Object System.Collections.ArrayList
    while ($true) {
        $process.Refresh()
        $treeStats = Get-CheckProcessTreeStats -RootPid $process.Id
        if ([int64]$treeStats.working_set_bytes -gt $peakTreeWorkingSetBytes) {
            $peakTreeWorkingSetBytes = [int64]$treeStats.working_set_bytes
        }

        if ($MonitorPath -ne '') {
            $size = [int64]0
            if ($MonitorPath.Contains('*') -or $MonitorPath.Contains('?')) {
                $matchedFiles = @(Get-ChildItem -Path $MonitorPath -File -ErrorAction SilentlyContinue)
                if ($matchedFiles.Count -gt 0) {
                    $size = [int64](@($matchedFiles | Measure-Object -Property Length -Sum)[0].Sum)
                    $samples.Add($size) | Out-Null
                }
            } elseif (Test-Path -LiteralPath $MonitorPath) {
                $size = [int64](Get-Item -LiteralPath $MonitorPath).Length
                $samples.Add($size) | Out-Null
            }
        }

        if ($process.HasExited) {
            break
        }
        $elapsed = (Get-Date) - $started
        if ($elapsed.TotalSeconds -ge $TimeoutSec) {
            $timedOut = $true
            Stop-CheckProcessTree -RootPid $process.Id
            break
        }
        Start-Sleep -Milliseconds 200
    }

    if (-not $timedOut) {
        $process.WaitForExit()
        $process.Refresh()
    }

    if ($MonitorPath -ne '') {
        if ($MonitorPath.Contains('*') -or $MonitorPath.Contains('?')) {
            $matchedFiles = @(Get-ChildItem -Path $MonitorPath -File -ErrorAction SilentlyContinue)
            if ($matchedFiles.Count -gt 0) {
                $size = [int64](@($matchedFiles | Measure-Object -Property Length -Sum)[0].Sum)
                $samples.Add($size) | Out-Null
            }
        } elseif (Test-Path -LiteralPath $MonitorPath) {
            $size = [int64](Get-Item -LiteralPath $MonitorPath).Length
            $samples.Add($size) | Out-Null
        }
    }

    $ended = Get-Date
    $exitCode = if ($timedOut) { 124 } else { [int]$process.ExitCode }
    if (-not [string]::IsNullOrWhiteSpace($ExitCodeSidecarPath) -and (Test-Path -LiteralPath $ExitCodeSidecarPath)) {
        $sidecarRaw = [string](Get-Content -LiteralPath $ExitCodeSidecarPath -Raw)
        $parsedSidecar = 0
        if ([int]::TryParse($sidecarRaw.Trim(), [ref]$parsedSidecar)) {
            $exitCode = $parsedSidecar
        }
    }
    $stdoutText = if (Test-Path -LiteralPath $stdoutPath) { Get-Content -LiteralPath $stdoutPath -Raw } else { '' }
    $stderrText = if (Test-Path -LiteralPath $stderrPath) { Get-Content -LiteralPath $stderrPath -Raw } else { '' }
    $stderrNonEmpty = -not [string]::IsNullOrWhiteSpace($stderrText)

    @(
        "name: $Name"
        "file: $FilePath"
        "arguments: $($Arguments -join ' ')"
        "working_directory: $WorkingDirectory"
        "timeout_sec: $TimeoutSec"
        "timed_out: $timedOut"
        "expected_failure: $ExpectedFailure"
        "peak_tree_working_set_bytes: $peakTreeWorkingSetBytes"
        "exit_code: $exitCode"
        "stderr_nonempty: $stderrNonEmpty"
        "started_at: $($started.ToString('o'))"
        "ended_at: $($ended.ToString('o'))"
        ""
        "stdout:"
        $stdoutText
        ""
        "stderr:"
        $stderrText
    ) | Set-Content -LiteralPath $logPath -Encoding UTF8

    $record = [ordered]@{
        name = $Name
        file = $FilePath
        arguments = @($Arguments)
        cwd = $WorkingDirectory
        working_directory = $WorkingDirectory
        timeout_sec = $TimeoutSec
        timed_out = $timedOut
        expected_failure = $ExpectedFailure
        peak_tree_working_set_bytes = $peakTreeWorkingSetBytes
        exit_code = $exitCode
        stderr_nonempty = $stderrNonEmpty
        duration_ms = [int](($ended - $started).TotalMilliseconds)
        started_at = $started.ToString('o')
        ended_at = $ended.ToString('o')
        stdout = "logs/$safeName.stdout.log"
        stderr = "logs/$safeName.stderr.log"
        log = "logs/$safeName.log"
        monitor_sizes = @($samples)
    }

    ($record | ConvertTo-Json -Compress) | Add-Content -LiteralPath $Ctx.CommandsPath -Encoding UTF8
    $Ctx.CommandResults[$Name] = $record
    return $record
}

function Get-CommandStdoutText {
    param(
        [Parameter(Mandatory=$true)]$Ctx,
        [Parameter(Mandatory=$true)]$Record
    )

    $path = Join-Path $Ctx.ResultDir ([string]$Record.stdout).Replace('/', '\')
    if (-not (Test-Path -LiteralPath $path)) {
        return ''
    }
    return [string](Get-Content -LiteralPath $path -Raw)
}

function Get-CommandStderrText {
    param(
        [Parameter(Mandatory=$true)]$Ctx,
        [Parameter(Mandatory=$true)]$Record
    )

    $path = Join-Path $Ctx.ResultDir ([string]$Record.stderr).Replace('/', '\')
    if (-not (Test-Path -LiteralPath $path)) {
        return ''
    }
    return [string](Get-Content -LiteralPath $path -Raw)
}

function Test-CommandSuccess {
    param(
        [Parameter(Mandatory=$true)]$Record,
        [bool]$AllowStderr = $false
    )

    $stderrOk = $AllowStderr -or (-not [bool]$Record.stderr_nonempty)
    return ((-not [bool]$Record.timed_out) -and ([int]$Record.exit_code -eq 0) -and $stderrOk)
}

function Invoke-CheckExpectedFailure {
    param(
        [Parameter(Mandatory=$true)]$Ctx,
        [Parameter(Mandatory=$true)][string]$Name,
        [Parameter(Mandatory=$true)][string]$ToolPath,
        [Parameter(Mandatory=$true)][string[]]$ToolArgs,
        [Parameter(Mandatory=$true)][string[]]$ExpectedSnippets
    )

    $escapedTool = $ToolPath.Replace("'", "''")
    $quotedArgs = @($ToolArgs | ForEach-Object { "'" + $_.Replace("'", "''") + "'" }) -join ', '
    $wrapperPath = Join-Path $Ctx.TmpDir ($Name + '.expected_failure.ps1')
    $sidecarPath = Join-Path $Ctx.TmpDir ($Name + '.expected_failure.exit_code.txt')
    $wrapperBody = @"
`$ErrorActionPreference = 'Stop'
& '$escapedTool' @($quotedArgs)
`$exitCode = `$LASTEXITCODE
if (`$null -eq `$exitCode) {
    `$exitCode = 0
}
Set-Content -LiteralPath '$($sidecarPath.Replace("'", "''"))' -Value ([string]`$exitCode) -Encoding UTF8
exit `$exitCode
"@
    Set-Content -LiteralPath $wrapperPath -Value $wrapperBody -Encoding UTF8
    $record = Invoke-CheckHiddenCommand -Ctx $Ctx -Name $Name -FilePath 'powershell.exe' -Arguments @('-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', $wrapperPath) -ExpectedFailure $true -ExitCodeSidecarPath $sidecarPath
    $output = (Get-CommandStdoutText -Ctx $Ctx -Record $record) + "`n" + (Get-CommandStderrText -Ctx $Ctx -Record $record)
    $snippetsOk = $true
    foreach ($snippet in $ExpectedSnippets) {
        if (-not $output.Contains($snippet)) {
            $snippetsOk = $false
            break
        }
    }
    $nonZero = ([int]$record.exit_code -ne 0)
    return [ordered]@{
        record = $record
        output = $output
        nonzero = $nonZero
        snippets_ok = $snippetsOk
        ok = ($nonZero -and $snippetsOk -and -not [bool]$record.timed_out)
    }
}

function Add-FeatureAssessment {
    param(
        [Parameter(Mandatory=$true)]$Ctx,
        [Parameter(Mandatory=$true)][string]$Id,
        [Parameter(Mandatory=$true)][ValidateSet('minimum','good','excellent','engineering')][string]$Level,
        [Parameter(Mandatory=$true)][string]$Category,
        [Parameter(Mandatory=$true)][string]$Requirement,
        [Parameter(Mandatory=$true)][ValidateSet('not_implemented','partial','full')][string]$Implementation,
        [Parameter(Mandatory=$true)][ValidateSet('not_tested','nonconformant','conformant')][string]$Conformance,
        [string[]]$Evidence = @(),
        [string]$Details = ''
    )

    $normalized = @($Evidence | Where-Object { -not [string]::IsNullOrWhiteSpace($_) } | ForEach-Object { [string]$_ } | Select-Object -Unique)
    if ($normalized.Count -eq 0) {
        $normalized = @('commands.jsonl')
    }

    $item = [ordered]@{
        id = $Id
        level = $Level
        category = $Category
        requirement = $Requirement
        implementation = $Implementation
        conformance = $Conformance
        evidence = $normalized
        details = $Details
    }
    $Ctx.Assessments.Add($item) | Out-Null
}

function Add-BooleanFeatureAssessment {
    param(
        [Parameter(Mandatory=$true)]$Ctx,
        [Parameter(Mandatory=$true)][string]$Id,
        [Parameter(Mandatory=$true)][ValidateSet('minimum','good','excellent','engineering')][string]$Level,
        [Parameter(Mandatory=$true)][string]$Category,
        [Parameter(Mandatory=$true)][string]$Requirement,
        [Parameter(Mandatory=$true)][bool]$Ok,
        [string[]]$Evidence = @(),
        [string]$Details = ''
    )

    $implementation = if ($Ok) { 'full' } else { 'partial' }
    $conformance = if ($Ok) { 'conformant' } else { 'nonconformant' }
    Add-FeatureAssessment -Ctx $Ctx -Id $Id -Level $Level -Category $Category -Requirement $Requirement -Implementation $implementation -Conformance $conformance -Evidence $Evidence -Details $Details
}

function Convert-CheckJsonl {
    param([string]$Path)

    if (-not (Test-Path -LiteralPath $Path)) {
        return @()
    }
    $items = New-Object System.Collections.ArrayList
    foreach ($line in Get-Content -LiteralPath $Path -Encoding UTF8) {
        if ([string]::IsNullOrWhiteSpace($line)) {
            continue
        }
        $items.Add(($line | ConvertFrom-Json -ErrorAction Stop)) | Out-Null
    }
    return @($items)
}

function Test-JsonlFinalLF {
    param([string]$Path)

    if (-not (Test-Path -LiteralPath $Path)) {
        return $false
    }
    $bytes = [System.IO.File]::ReadAllBytes($Path)
    if ($bytes.Length -eq 0) {
        return $false
    }
    return ($bytes[$bytes.Length - 1] -eq 10)
}

function Get-CheckLineCount {
    param([string]$Path)

    if (-not (Test-Path -LiteralPath $Path)) {
        return 0
    }
    $count = 0
    $reader = New-Object System.IO.StreamReader($Path)
    try {
        while (($null -ne $reader.ReadLine())) {
            $count++
        }
    } finally {
        $reader.Close()
    }
    return $count
}

function New-LargeFixtures {
    param(
        [Parameter(Mandatory=$true)][string]$ScenarioPath,
        [Parameter(Mandatory=$true)][string]$PoliciesPath,
        [Parameter(Mandatory=$true)][int]$EventCount,
        [Parameter(Mandatory=$true)][int]$PolicyCount
    )

    $utf8 = New-Object System.Text.UTF8Encoding($false)
    $writer = New-Object System.IO.StreamWriter($ScenarioPath, $false, $utf8)
    try {
        $writer.WriteLine('scenario_id: large_100k')
        $writer.WriteLine('name: Large scenario')
        $writer.WriteLine('users:')
        $writer.WriteLine('  - user_id: user_001')
        $writer.WriteLine('    department: sec')
        $writer.WriteLine('    role: analyst')
        $writer.WriteLine('events:')
        for ($index = 1; $index -le $EventCount; $index++) {
            $writer.WriteLine("  - event_id: evt_$index")
            $writer.WriteLine('    time: "10:00"')
            $writer.WriteLine('    user_id: user_001')
            $writer.WriteLine('    action: match_action')
        }
    } finally {
        $writer.Flush()
        $writer.Close()
    }

    $policyWriter = New-Object System.IO.StreamWriter($PoliciesPath, $false, $utf8)
    try {
        $policyWriter.WriteLine('policies:')
        $policyWriter.WriteLine('  - policy_id: p_match')
        $policyWriter.WriteLine('    name: Match')
        $policyWriter.WriteLine('    severity: low')
        $policyWriter.WriteLine('    condition:')
        $policyWriter.WriteLine('      field: action')
        $policyWriter.WriteLine('      equals: match_action')
        for ($index = 1; $index -lt $PolicyCount; $index++) {
            $policyWriter.WriteLine("  - policy_id: p_miss_$index")
            $policyWriter.WriteLine('    name: Miss')
            $policyWriter.WriteLine('    severity: low')
            $policyWriter.WriteLine('    condition:')
            $policyWriter.WriteLine('      field: action')
            $policyWriter.WriteLine("      equals: never_$index")
        }
    } finally {
        $policyWriter.Flush()
        $policyWriter.Close()
    }
}

function Copy-CheckSnapshot {
    param([Parameter(Mandatory=$true)]$Ctx)

    foreach ($name in @('README.md', 'Makefile', 'go.mod', 'docs')) {
        $source = Join-Path $Ctx.RepoRoot $name
        if (-not (Test-Path -LiteralPath $source)) {
            continue
        }
        $destination = Join-Path $Ctx.ResultDir (Join-Path 'repo_snapshot' $name)
        $parent = Split-Path -Parent $destination
        New-Item -ItemType Directory -Force -Path $parent | Out-Null
        Copy-Item -LiteralPath $source -Destination $destination -Recurse -Force
    }
}

function Complete-Check {
    param(
        [Parameter(Mandatory=$true)]$Ctx,
        [hashtable]$Extra = @{}
    )

    $expectedByLevel = [ordered]@{
        minimum = 6
        good = 5
        excellent = 4
        engineering = 12
    }
    $assessmentItems = @($Ctx.Assessments)
    $ids = @($assessmentItems | ForEach-Object { $_.id })
    $uniqueIds = @($ids | Select-Object -Unique)
    if ($ids.Count -ne $uniqueIds.Count) {
        throw "duplicate assessment ids detected: total=$($ids.Count) unique=$($uniqueIds.Count)"
    }
    foreach ($level in $expectedByLevel.Keys) {
        $actualCount = @($assessmentItems | Where-Object { $_.level -eq $level }).Count
        if ($actualCount -ne [int]$expectedByLevel[$level]) {
            throw "unexpected feature count for ${level}: got=$actualCount expected=$($expectedByLevel[$level])"
        }
    }
    if ($assessmentItems.Count -ne 27) {
        throw "unexpected total feature count: $($assessmentItems.Count)"
    }

    $summary = [ordered]@{}
    foreach ($level in @('minimum', 'good', 'excellent', 'engineering')) {
        $items = @($assessmentItems | Where-Object { $_.level -eq $level })
        $summary[$level] = [ordered]@{
            total = $items.Count
            full = @($items | Where-Object { $_.implementation -eq 'full' }).Count
            partial = @($items | Where-Object { $_.implementation -eq 'partial' }).Count
            not_implemented = @($items | Where-Object { $_.implementation -eq 'not_implemented' }).Count
            conformant = @($items | Where-Object { $_.conformance -eq 'conformant' }).Count
            nonconformant = @($items | Where-Object { $_.conformance -eq 'nonconformant' }).Count
            not_tested = @($items | Where-Object { $_.conformance -eq 'not_tested' }).Count
        }
    }

    Save-CheckJson -Path (Join-Path $Ctx.ResultDir 'assessment.json') -Value ([ordered]@{
        schema_version = 1
        statuses = [ordered]@{
            implementation = @('not_implemented', 'partial', 'full')
            conformance = @('not_tested', 'nonconformant', 'conformant')
        }
        summary = $summary
        features = $assessmentItems
    })

    Copy-CheckSnapshot -Ctx $Ctx

    $commands = @($Ctx.CommandResults.Values)
    $expectedFailureUnexpected = @($commands | Where-Object { $_.expected_failure -and ([int]$_.exit_code -eq 0 -or [bool]$_.timed_out) })
    $unexpectedFailures = @($commands | Where-Object { (-not $_.expected_failure) -and (([int]$_.exit_code -ne 0) -or [bool]$_.timed_out -or [bool]$_.stderr_nonempty) })
    $failedCommands = @($commands | Where-Object { ([int]$_.exit_code -ne 0) -or [bool]$_.timed_out })
    $timedOut = @($commands | Where-Object { [bool]$_.timed_out })
    $selfPath = $PSCommandPath
    $selfHash = ''
    if (-not [string]::IsNullOrWhiteSpace($selfPath) -and (Test-Path -LiteralPath $selfPath)) {
        $sha = [System.Security.Cryptography.SHA256]::Create()
        $stream = [System.IO.File]::OpenRead($selfPath)
        try {
            $hashBytes = $sha.ComputeHash($stream)
            $selfHash = ([System.BitConverter]::ToString($hashBytes)).Replace('-', '').ToLowerInvariant()
        } finally {
            $stream.Dispose()
            $sha.Dispose()
        }
    }

    Save-CheckJson -Path (Join-Path $Ctx.ResultDir 'manifest.json') -Value ([ordered]@{
        student = 'policy_playground_check'
        repo_root = $Ctx.RepoRoot
        started_at = $Ctx.StartedAt
        completed_at = (Get-Date).ToString('o')
        result_dir = $Ctx.ResultDir
        commands_file = 'commands.jsonl'
        assessment_file = 'assessment.json'
        notes = [ordered]@{
            checker_sha256 = $selfHash
            command_count = $commands.Count
            failed_commands_total = $failedCommands.Count
            failed_commands_expected = @($commands | Where-Object { $_.expected_failure -and ([int]$_.exit_code -ne 0) -and (-not [bool]$_.timed_out) }).Count
            failed_commands_unexpected = ($expectedFailureUnexpected.Count + $unexpectedFailures.Count)
            expected_failure_unexpected_zero_exit = $expectedFailureUnexpected.Count
            timed_out_commands = $timedOut.Count
            extra = $Extra
        }
    })

    $zipPath = "$($Ctx.ResultDir).zip"
    if (Test-Path -LiteralPath $zipPath) {
        Remove-Item -LiteralPath $zipPath -Force
    }
    Compress-Archive -Path (Join-Path $Ctx.ResultDir '*') -DestinationPath $zipPath -Force
    Write-Host "CHECK_RESULT_DIR=$($Ctx.ResultDir)"
    Write-Host "CHECK_RESULT_ZIP=$zipPath"
    return $zipPath
}

$ctx = New-CheckContext -Root $RepoRoot -OutRoot $OutRoot
$toolPath = Join-Path $ctx.OutputsDir 'policy-playground.exe'

foreach ($stale in @(
    (Join-Path $ctx.RepoRoot '.tmp_bad_policies.yaml'),
    (Join-Path $ctx.RepoRoot '.check-hidden-stdout.log'),
    (Join-Path $ctx.RepoRoot '.check-hidden-stderr.log')
)) {
    if (Test-Path -LiteralPath $stale) {
        Remove-Item -LiteralPath $stale -Force -ErrorAction SilentlyContinue
    }
}

$coreScenario = Write-CheckText -Ctx $ctx -RelativePath 'inputs/core/scenario.yaml' -Content @"
scenario_id: core_001
name: Core fixture
users:
  - user_id: user_001
    department: sales
    role: manager
events:
  - event_id: evt_001
    time: "10:00"
    user_id: user_001
    action: open_file
    object_type: file
    file_name: client_base.xlsx
    file_ext: xlsx
    content_classes: [client_data]
    channel: local
    destination_type: none
  - event_id: evt_002
    time: "10:05"
    user_id: user_001
    action: email_send
    object_type: file
    file_name: client_base.xlsx
    file_ext: xlsx
    content_classes: [client_data]
    channel: email
    destination_type: external
  - event_id: "evt_bad|<&tag>"
    time: "10:06"
    user_id: user_001
    action: open_file
    object_type: file
    file_name: public.txt
    file_ext: txt
    content_classes: [general]
    channel: local
    destination_type: none
"@

$corePolicies = Write-CheckText -Ctx $ctx -RelativePath 'inputs/core/policies.yaml' -Content @"
policies:
  - policy_id: p_all
    name: External data
    severity: high
    condition:
      all:
        - field: action
          equals: email_send
        - field: destination_type
          equals: external
        - field: file_ext
          in: [xlsx, docx, pdf]
        - field: content_classes
          contains: client_data
  - policy_id: p_any
    name: Never
    severity: low
    condition:
      any:
        - field: action
          equals: failed_login
        - field: channel
          equals: usb
"@

$anyScenario = Write-CheckText -Ctx $ctx -RelativePath 'inputs/any/scenario.yaml' -Content @"
scenario_id: any_001
users:
  - user_id: user_001
events:
  - event_id: any_001
    time: "10:00"
    user_id: user_001
    action: upload
    channel: chat
  - event_id: any_002
    time: "10:01"
    user_id: user_001
    action: download
    channel: email
  - event_id: any_003
    time: "10:02"
    user_id: user_001
    action: download
    channel: chat
"@

$anyPolicies = Write-CheckText -Ctx $ctx -RelativePath 'inputs/any/policies.yaml' -Content @"
policies:
  - policy_id: p_any
    name: Any branch
    severity: low
    condition:
      any:
        - field: action
          equals: upload
        - field: channel
          equals: email
"@

$existsScenario = Write-CheckText -Ctx $ctx -RelativePath 'inputs/exists/scenario.yaml' -Content @"
scenario_id: exists_001
users:
  - user_id: user_001
events:
  - event_id: ex_present
    time: "10:00"
    user_id: user_001
    channel: email
  - event_id: ex_missing
    time: "10:01"
    user_id: user_001
"@

$existsPolicies = Write-CheckText -Ctx $ctx -RelativePath 'inputs/exists/policies.yaml' -Content @"
policies:
  - policy_id: p_exists_true
    name: Exists true
    severity: low
    condition:
      field: channel
      exists: true
  - policy_id: p_exists_false
    name: Exists false
    severity: low
    condition:
      field: channel
      exists: false
"@

$windowScenario = Write-CheckText -Ctx $ctx -RelativePath 'inputs/window/scenario.yaml' -Content @"
scenario_id: win_001
users:
  - user_id: u1
  - user_id: u2
events:
  - event_id: w1
    time: "2024-01-01T10:00:00Z"
    user_id: u1
    action: failed_login
  - event_id: w2
    time: "2024-01-01T10:05:00Z"
    user_id: u1
    action: failed_login
  - event_id: w3
    time: "2024-01-01T10:10:00Z"
    user_id: u1
    action: failed_login
  - event_id: w5
    time: "2024-01-01T10:11:00Z"
    user_id: u2
    action: failed_login
  - event_id: w4
    time: "2024-01-01T10:20:01Z"
    user_id: u1
    action: failed_login
"@

$windowPolicies = Write-CheckText -Ctx $ctx -RelativePath 'inputs/window/policies.yaml' -Content @"
policies:
  - policy_id: p_window
    name: Window policy
    severity: medium
    condition:
      window:
        count:
          greater_than: 2
        within: 10m
        by: user_id
        where:
          field: action
          equals: failed_login
"@

$malformedScenario = Write-CheckText -Ctx $ctx -RelativePath 'inputs/invalid/malformed.yaml' -Content "scenario_id: bad`nusers:`n`t- user_id: x`nevents: []`n"
$unknownPolicies = Write-CheckText -Ctx $ctx -RelativePath 'inputs/invalid/unknown_policies.yaml' -Content @"
policies:
  - policy_id: p1
    name: bad
    unknown: true
    condition:
      field: action
      equals: open
"@
$duplicatePolicies = Write-CheckText -Ctx $ctx -RelativePath 'inputs/invalid/duplicate_policies.yaml' -Content @"
policies:
  - policy_id: p1
    name: first
    name: second
    condition:
      field: action
      equals: open
"@
$missingScenario = Write-CheckText -Ctx $ctx -RelativePath 'inputs/invalid/missing.yaml' -Content @"
scenario_id: bad
users:
  - department: sec
events:
  - event_id: e1
    time: "10:00"
    user_id: u1
"@

$metaHead = Invoke-CheckHiddenCommand -Ctx $ctx -Name 'meta_git_head' -FilePath $ctx.GitCmd -Arguments @('rev-parse', 'HEAD')
$metaStatus = Invoke-CheckHiddenCommand -Ctx $ctx -Name 'meta_git_status' -FilePath $ctx.GitCmd -Arguments @('status', '--short')
$metaGoVersion = Invoke-CheckHiddenCommand -Ctx $ctx -Name 'meta_go_version' -FilePath $ctx.GoCmd -Arguments @('version')
$metaGoEnv = Invoke-CheckHiddenCommand -Ctx $ctx -Name 'meta_go_env' -FilePath $ctx.GoCmd -Arguments @('env', 'GOVERSION', 'GOOS', 'GOARCH')

Set-Content -LiteralPath (Join-Path $ctx.MetaDir 'git_head.txt') -Value ((Get-CommandStdoutText -Ctx $ctx -Record $metaHead) -replace "`r?`n$", '') -Encoding UTF8
Set-Content -LiteralPath (Join-Path $ctx.MetaDir 'go_version.txt') -Value (Get-CommandStdoutText -Ctx $ctx -Record $metaGoVersion) -Encoding UTF8
Set-Content -LiteralPath (Join-Path $ctx.MetaDir 'go_env.txt') -Value (Get-CommandStdoutText -Ctx $ctx -Record $metaGoEnv) -Encoding UTF8
Set-Content -LiteralPath (Join-Path $ctx.MetaDir 'git_status_short.txt') -Value ((Get-CommandStdoutText -Ctx $ctx -Record $metaStatus).TrimEnd()) -Encoding UTF8

$goFmt = Invoke-CheckHiddenCommand -Ctx $ctx -Name 'go_fmt' -FilePath $ctx.GoCmd -Arguments @('fmt', './...')
$goTestAll = Invoke-CheckHiddenCommand -Ctx $ctx -Name 'go_test_all' -FilePath $ctx.GoCmd -Arguments @('test', './...')
$makeTest = $null
$makeBench = $null
$makeDemo = $null
if ($ctx.MakeCmd -ne '') {
    $makeTest = Invoke-CheckHiddenCommand -Ctx $ctx -Name 'make_test' -FilePath $ctx.MakeCmd -Arguments @('test')
    $makeBench = Invoke-CheckHiddenCommand -Ctx $ctx -Name 'make_bench' -FilePath $ctx.MakeCmd -Arguments @('bench')
    $makeDemo = Invoke-CheckHiddenCommand -Ctx $ctx -Name 'make_demo' -FilePath $ctx.MakeCmd -Arguments @('demo')
}

$conditionJson = Invoke-CheckHiddenCommand -Ctx $ctx -Name 'go_test_condition_json' -FilePath $ctx.GoCmd -Arguments @('test', './internal/playground', '-run', '^TestCondition(Equals|In|Contains|All)$', '-count=1', '-json')
$targetedTests = Invoke-CheckHiddenCommand -Ctx $ctx -Name 'go_test_targeted' -FilePath $ctx.GoCmd -Arguments @('test', './internal/playground', '-run', '^(TestConditionAnyBothBranchesAndFalse|TestConditionExistsTrueFalsePresentMissing|TestWindowInclusiveBoundaryByUser|TestLoadScenarioMalformedErrorWithCoordinatesAndPath|TestLoadScenarioUnknownKey|TestLoadScenarioMissingRequired|TestLoadPoliciesUnknownField|TestLoadPoliciesDuplicateKey|TestLoadScenarioQuotedHostileScalar|TestExecuteExplainJSONIncludesFailureOperator)$', '-count=1')
$benchRun = Invoke-CheckHiddenCommand -Ctx $ctx -Name 'go_bench_single' -FilePath $ctx.GoCmd -Arguments @('test', './cmd/policy-playground', '-run', '^$', '-bench', 'BenchmarkPolicy100000Events100Policies', '-benchmem', '-benchtime=1x', '-count=1')
$buildCli = Invoke-CheckHiddenCommand -Ctx $ctx -Name 'build_cli' -FilePath $ctx.GoCmd -Arguments @('build', '-o', $toolPath, './cmd/policy-playground')

$coreOut = Join-Path $ctx.OutputsDir 'core_alerts.jsonl'
$coreReport = Join-Path $ctx.OutputsDir 'core_report.md'
$coreHTML = Join-Path $ctx.OutputsDir 'core_report.html'
$coreRun = Invoke-CheckHiddenCommand -Ctx $ctx -Name 'cli_run_core' -FilePath $toolPath -Arguments @('run', '--scenario', $coreScenario, '--policies', $corePolicies, '--out', $coreOut, '--report', $coreReport, '--html', $coreHTML)
$coreExplain = Invoke-CheckHiddenCommand -Ctx $ctx -Name 'cli_explain_evt002' -FilePath $toolPath -Arguments @('explain', '--scenario', $coreScenario, '--policies', $corePolicies, '--event', 'evt_002', '--show-not-matched')
$coreExplainJSON = Invoke-CheckHiddenCommand -Ctx $ctx -Name 'cli_explain_evt002_json' -FilePath $toolPath -Arguments @('explain', '--scenario', $coreScenario, '--policies', $corePolicies, '--event', 'evt_002', '--show-not-matched', '--format', 'json')
$coreExplainNoMatch = Invoke-CheckHiddenCommand -Ctx $ctx -Name 'cli_explain_evt001' -FilePath $toolPath -Arguments @('explain', '--scenario', $coreScenario, '--policies', $corePolicies, '--event', 'evt_001')
$coreExplainUnknown = Invoke-CheckExpectedFailure -Ctx $ctx -Name 'cli_explain_unknown' -ToolPath $toolPath -ToolArgs @('explain', '--scenario', $coreScenario, '--policies', $corePolicies, '--event', 'evt_missing') -ExpectedSnippets @('not found')

$anyOut = Join-Path $ctx.OutputsDir 'any_alerts.jsonl'
$anyRun = Invoke-CheckHiddenCommand -Ctx $ctx -Name 'cli_run_any' -FilePath $toolPath -Arguments @('run', '--scenario', $anyScenario, '--policies', $anyPolicies, '--out', $anyOut)

$existsOut = Join-Path $ctx.OutputsDir 'exists_alerts.jsonl'
$existsRun = Invoke-CheckHiddenCommand -Ctx $ctx -Name 'cli_run_exists' -FilePath $toolPath -Arguments @('run', '--scenario', $existsScenario, '--policies', $existsPolicies, '--out', $existsOut)

$invalidMalformed = Invoke-CheckExpectedFailure -Ctx $ctx -Name 'cli_invalid_malformed' -ToolPath $toolPath -ToolArgs @('run', '--scenario', $malformedScenario, '--policies', $corePolicies, '--out', (Join-Path $ctx.OutputsDir 'invalid_malformed.jsonl')) -ExpectedSnippets @('malformed.yaml', '$')
$invalidUnknown = Invoke-CheckExpectedFailure -Ctx $ctx -Name 'cli_invalid_unknown' -ToolPath $toolPath -ToolArgs @('run', '--scenario', $coreScenario, '--policies', $unknownPolicies, '--out', (Join-Path $ctx.OutputsDir 'invalid_unknown.jsonl')) -ExpectedSnippets @('unknown_policies.yaml', '$.policies[0].unknown')
$invalidDuplicate = Invoke-CheckExpectedFailure -Ctx $ctx -Name 'cli_invalid_duplicate' -ToolPath $toolPath -ToolArgs @('run', '--scenario', $coreScenario, '--policies', $duplicatePolicies, '--out', (Join-Path $ctx.OutputsDir 'invalid_duplicate.jsonl')) -ExpectedSnippets @('duplicate_policies.yaml', 'дублирующийся ключ', '$.policies[0].name')
$invalidMissing = Invoke-CheckExpectedFailure -Ctx $ctx -Name 'cli_invalid_missing' -ToolPath $toolPath -ToolArgs @('run', '--scenario', $missingScenario, '--policies', $corePolicies, '--out', (Join-Path $ctx.OutputsDir 'invalid_missing.jsonl')) -ExpectedSnippets @('missing.yaml', '$.users[0].user_id')

$windowOut = Join-Path $ctx.OutputsDir 'window_alerts.jsonl'
$windowRun = Invoke-CheckHiddenCommand -Ctx $ctx -Name 'cli_run_window' -FilePath $toolPath -Arguments @('run', '--scenario', $windowScenario, '--policies', $windowPolicies, '--out', $windowOut)

$largeScenario = Join-Path $ctx.InputsDir 'large/scenario.yaml'
$largePolicies = Join-Path $ctx.InputsDir 'large/policies.yaml'
$largeOut = Join-Path $ctx.OutputsDir 'large_alerts.jsonl'
New-Item -ItemType Directory -Force -Path (Split-Path -Parent $largeScenario) | Out-Null
New-LargeFixtures -ScenarioPath $largeScenario -PoliciesPath $largePolicies -EventCount 100000 -PolicyCount 100
$largeMonitorPath = Join-Path (Split-Path -Parent $largeOut) 'large_alerts.jsonl*'
$largeRun = Invoke-CheckHiddenCommand -Ctx $ctx -Name 'cli_run_large' -FilePath $toolPath -Arguments @('run', '--scenario', $largeScenario, '--policies', $largePolicies, '--out', $largeOut) -MonitorPath $largeMonitorPath
$largeAlertCount = Get-CheckLineCount -Path $largeOut
$largeDurationSec = [math]::Round(([double]$largeRun.duration_ms / 1000.0), 3)
$largeSamples = @($largeRun.monitor_sizes)
$largeGrowthCheckpoints = 0
for ($index = 1; $index -lt $largeSamples.Count; $index++) {
    if ([int64]$largeSamples[$index] -gt [int64]$largeSamples[$index - 1]) {
        $largeGrowthCheckpoints++
    }
}
$largeGrowthOk = ($largeGrowthCheckpoints -ge 2)

$coreJsonlParseOk = $true
$coreAlerts = @()
try {
    $coreAlerts = @(Convert-CheckJsonl -Path $coreOut)
} catch {
    $coreJsonlParseOk = $false
}
$coreLineCount = Get-CheckLineCount -Path $coreOut
$coreFinalLF = Test-JsonlFinalLF -Path $coreOut
$coreFieldsOk = $false
$coreMatchedFieldOk = $false
if ($coreAlerts.Count -gt 0) {
    $expectedFields = @('event_id', 'matched', 'policy_id', 'policy_name', 'reasons', 'severity', 'user_id') | Sort-Object
    $actualFields = @($coreAlerts[0].PSObject.Properties.Name) | Sort-Object
    $coreFieldsOk = (($expectedFields -join '|') -eq ($actualFields -join '|'))
    $coreMatchedFieldOk = (@($coreAlerts | Where-Object { $_.matched -ne $true }).Count -eq 0)
}

$coreExact = $false
$coreReasonsOk = $false
$coreEvt001NoAlert = $false
if ($coreAlerts.Count -eq 1) {
    $alert = $coreAlerts[0]
    $coreExact = ($alert.policy_id -eq 'p_all' -and $alert.event_id -eq 'evt_002' -and $alert.matched -eq $true)
    $coreReasonsOk = (($alert.reasons -join '|') -eq 'action equals email_send|destination_type equals external|file_ext in [xlsx, docx, pdf]|content_classes contains client_data')
}
$coreEvt001NoAlert = (@($coreAlerts | Where-Object { $_.event_id -eq 'evt_001' }).Count -eq 0)
$coreJsonlOk = $coreJsonlParseOk -and ($coreLineCount -eq $coreAlerts.Count) -and $coreFinalLF -and $coreFieldsOk -and $coreMatchedFieldOk

$anyAlerts = @()
$anyParseOk = $true
try {
    $anyAlerts = @(Convert-CheckJsonl -Path $anyOut)
} catch {
    $anyParseOk = $false
}
$anyMap = @{}
foreach ($alert in $anyAlerts) {
    $anyMap[[string]$alert.event_id] = [string]($alert.reasons -join '|')
}
$anyOk = (Test-CommandSuccess -Record $anyRun) -and $anyParseOk -and ($anyAlerts.Count -eq 2) -and ($anyMap.ContainsKey('any_001')) -and ($anyMap.ContainsKey('any_002')) -and ($anyMap['any_001'] -eq 'action equals upload') -and ($anyMap['any_002'] -eq 'channel equals email') -and (-not $anyMap.ContainsKey('any_003'))

$existsAlerts = @()
$existsParseOk = $true
try {
    $existsAlerts = @(Convert-CheckJsonl -Path $existsOut)
} catch {
    $existsParseOk = $false
}
$existsPairs = @($existsAlerts | ForEach-Object { "$($_.policy_id)|$($_.event_id)" })
$existsOk = (Test-CommandSuccess -Record $existsRun) -and $existsParseOk -and ($existsAlerts.Count -eq 2) -and ($existsPairs -contains 'p_exists_true|ex_present') -and ($existsPairs -contains 'p_exists_false|ex_missing')

$windowAlerts = @()
$windowParseOk = $true
try {
    $windowAlerts = @(Convert-CheckJsonl -Path $windowOut)
} catch {
    $windowParseOk = $false
}
$windowOk = (Test-CommandSuccess -Record $windowRun) -and $windowParseOk -and ($windowAlerts.Count -eq 1) -and ($windowAlerts[0].event_id -eq 'w3')

$conditionPass = $false
$conditionSuiteNonEmpty = $false
$conditionStdout = Get-CommandStdoutText -Ctx $ctx -Record $conditionJson
$conditionLines = @($conditionStdout -split "`r?`n" | Where-Object { $_ -like '{*' })
$passed = @{}
foreach ($line in $conditionLines) {
    try {
        $obj = $line | ConvertFrom-Json -ErrorAction Stop
        if (($obj.PSObject.Properties.Name -contains 'Action') -and ($obj.PSObject.Properties.Name -contains 'Test')) {
            if ($obj.Action -eq 'pass' -and -not [string]::IsNullOrWhiteSpace([string]$obj.Test)) {
                $passed[[string]$obj.Test] = $true
            }
        }
    } catch {
    }
}
$conditionSuiteNonEmpty = ($conditionLines.Count -gt 0)
$conditionPass = $passed.ContainsKey('TestConditionEquals') -and $passed.ContainsKey('TestConditionIn') -and $passed.ContainsKey('TestConditionContains') -and $passed.ContainsKey('TestConditionAll')

$reportText = if (Test-Path -LiteralPath $coreReport) { [string](Get-Content -LiteralPath $coreReport -Raw) } else { '' }
$htmlText = if (Test-Path -LiteralPath $coreHTML) { [string](Get-Content -LiteralPath $coreHTML -Raw) } else { '' }
$explainText = Get-CommandStdoutText -Ctx $ctx -Record $coreExplain
$explainNoMatchText = Get-CommandStdoutText -Ctx $ctx -Record $coreExplainNoMatch
$explainJSONText = Get-CommandStdoutText -Ctx $ctx -Record $coreExplainJSON
$benchText = Get-CommandStdoutText -Ctx $ctx -Record $benchRun

$markdownOk = $reportText.Contains('policy-playground') -and $reportText.Contains('| p_all | evt_002 |') -and $reportText.Contains('action equals email_send') -and $reportText.Contains('evt_bad\|&lt;&amp;tag&gt;')
$htmlOk = $htmlText.Contains('<table>') -and $htmlText.Contains('evt_bad|&lt;&amp;tag&gt;') -and (-not $htmlText.Contains('evt_bad|<&tag>')) -and $htmlText.Contains('action equals email_send')
$reportSemanticsOk = $reportText.Contains('| p_all | evt_002 | user_001 | action equals email_send; destination_type equals external; file_ext in [xlsx, docx, pdf]; content_classes contains client_data |') -and $htmlText.Contains('<td>p_all</td><td>evt_002</td><td>user_001</td>')

$explainMatchedOk = $explainText.Contains('action equals email_send') -and $explainText.Contains('destination_type equals external')
$explainNoMatchOk = $explainNoMatchText.Contains('matched policies:') -and $explainNoMatchText.Contains('- none')
$explainUnknownOk = $coreExplainUnknown.ok -and $coreExplainUnknown.nonzero
$goodExplainOk = (Test-CommandSuccess -Record $coreExplain) -and (Test-CommandSuccess -Record $coreExplainNoMatch) -and $explainMatchedOk -and $explainNoMatchOk -and $explainUnknownOk

$whyNotStructured = $false
if (Test-CommandSuccess -Record $coreExplainJSON) {
    try {
        $explainObj = $explainJSONText | ConvertFrom-Json -ErrorAction Stop
        if ($null -ne $explainObj.not_matched -and $explainObj.not_matched.Count -gt 0 -and $explainObj.not_matched[0].failures.Count -gt 0) {
            $failure = $explainObj.not_matched[0].failures[0]
            $names = @($failure.PSObject.Properties.Name)
            $whyNotStructured = ($names -contains 'path') -and ($names -contains 'field') -and ($names -contains 'operator') -and ($names -contains 'expected') -and ($names -contains 'actual') -and ($names -contains 'present')
        }
    } catch {
        $whyNotStructured = $false
    }
}

$yamlCoordPattern = ':[0-9]+:[0-9]+\s+\$.+'
$yamlRuntimeOk = $invalidMalformed.ok -and $invalidUnknown.ok -and $invalidDuplicate.ok -and $invalidMissing.ok -and ($invalidMalformed.output -match $yamlCoordPattern) -and ($invalidUnknown.output -match $yamlCoordPattern) -and ($invalidDuplicate.output -match $yamlCoordPattern) -and ($invalidMissing.output -match $yamlCoordPattern)

$benchOk = (Test-CommandSuccess -Record $benchRun) -and ($benchText -match 'BenchmarkPolicy100000Events100Policies') -and ($benchText -match 'ns/op') -and ($benchText -match 'B/op') -and ($benchText -match 'allocs/op')
$largeOk = (Test-CommandSuccess -Record $largeRun) -and (-not [bool]$largeRun.timed_out) -and ($largeDurationSec -le 10) -and ($largeAlertCount -eq 100000) -and $largeGrowthOk -and ([int64]$largeRun.peak_tree_working_set_bytes -gt 0)

$cleanupTargets = @(
    $largeScenario,
    $largePolicies,
    $largeOut,
    (Join-Path $ctx.RepoRoot '.tmp_bad_policies.yaml')
)
$cleanupRemoved = New-Object System.Collections.ArrayList
$cleanupOk = $true
foreach ($target in $cleanupTargets) {
    if (Test-Path -LiteralPath $target) {
        try {
            Remove-Item -LiteralPath $target -Force -Recurse -ErrorAction Stop
            $cleanupRemoved.Add($target) | Out-Null
        } catch {
            $cleanupOk = $false
        }
    }
}

Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'large_metrics.json') -Value ([ordered]@{
    expected_evaluations = 10000000
    actual_alerts = $largeAlertCount
    expected_alerts = 100000
    external_wall_seconds = $largeDurationSec
    timeout_sec = $largeRun.timeout_sec
    timed_out = $largeRun.timed_out
    peak_tree_working_set_bytes = $largeRun.peak_tree_working_set_bytes
    output_size_samples = @($largeSamples)
    output_growth_checkpoints = $largeGrowthCheckpoints
    output_growth_ok = $largeGrowthOk
    cleanup_ok = $cleanupOk
})

$artifactPresence = [ordered]@{
    core_alerts_jsonl = (Test-Path -LiteralPath $coreOut)
    core_report_md = (Test-Path -LiteralPath $coreReport)
    core_report_html = (Test-Path -LiteralPath $coreHTML)
    any_alerts_jsonl = (Test-Path -LiteralPath $anyOut)
    exists_alerts_jsonl = (Test-Path -LiteralPath $existsOut)
    window_alerts_jsonl = (Test-Path -LiteralPath $windowOut)
    large_metrics_json = (Test-Path -LiteralPath (Join-Path $ctx.OutputsDir 'large_metrics.json'))
    git_status_short = (Test-Path -LiteralPath (Join-Path $ctx.MetaDir 'git_status_short.txt'))
}
Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'artifact_presence.json') -Value $artifactPresence

Add-BooleanFeatureAssessment -Ctx $ctx -Id 'minimum.read_scenario' -Level 'minimum' -Category 'input' -Requirement 'Read scenario YAML' -Ok ((Test-CommandSuccess -Record $coreRun) -and (-not [bool]$coreRun.stderr_nonempty)) -Evidence @([string]$coreRun.log)
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'minimum.read_policies' -Level 'minimum' -Category 'input' -Requirement 'Read policies YAML' -Ok ((Test-CommandSuccess -Record $coreRun) -and (-not [bool]$coreRun.stderr_nonempty)) -Evidence @([string]$coreRun.log)
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'minimum.conditions' -Level 'minimum' -Category 'algorithm' -Requirement 'equals in contains conditions' -Ok ($coreReasonsOk -and $coreEvt001NoAlert) -Evidence @([string]$coreRun.log, (Get-RelativeResultPath -Ctx $ctx -Path $coreOut))
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'minimum.all_group' -Level 'minimum' -Category 'algorithm' -Requirement 'all logical group' -Ok ($coreReasonsOk -and $coreEvt001NoAlert) -Evidence @([string]$coreRun.log, (Get-RelativeResultPath -Ctx $ctx -Path $coreOut))
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'minimum.alerts_jsonl' -Level 'minimum' -Category 'format' -Requirement 'JSONL alerts with reasons' -Ok ($coreExact -and $coreJsonlOk) -Evidence @((Get-RelativeResultPath -Ctx $ctx -Path $coreOut), [string]$coreRun.log)
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'minimum.condition_tests' -Level 'minimum' -Category 'tests' -Requirement 'Condition unit tests for equals in contains and all exist' -Ok ((Test-CommandSuccess -Record $conditionJson) -and $conditionSuiteNonEmpty -and $conditionPass) -Evidence @([string]$conditionJson.log)

Add-BooleanFeatureAssessment -Ctx $ctx -Id 'good.any_group' -Level 'good' -Category 'algorithm' -Requirement 'any logical group' -Ok ($anyOk -and (Test-CommandSuccess -Record $targetedTests)) -Evidence @([string]$anyRun.log, (Get-RelativeResultPath -Ctx $ctx -Path $anyOut), [string]$targetedTests.log)
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'good.exists' -Level 'good' -Category 'algorithm' -Requirement 'exists condition' -Ok ($existsOk -and (Test-CommandSuccess -Record $targetedTests)) -Evidence @([string]$existsRun.log, (Get-RelativeResultPath -Ctx $ctx -Path $existsOut), [string]$targetedTests.log)
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'good.markdown_report' -Level 'good' -Category 'format' -Requirement 'Markdown report' -Ok ($markdownOk -and $reportSemanticsOk) -Evidence @((Get-RelativeResultPath -Ctx $ctx -Path $coreReport), [string]$coreRun.log)
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'good.explain' -Level 'good' -Category 'cli' -Requirement 'explain command for one event' -Ok $goodExplainOk -Evidence @([string]$coreExplain.log, [string]$coreExplainNoMatch.log, [string]$coreExplainUnknown.record.log)
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'good.yaml_error_tests' -Level 'good' -Category 'tests' -Requirement 'YAML error tests with diagnostics' -Ok ($yamlRuntimeOk -and (Test-CommandSuccess -Record $targetedTests)) -Evidence @([string]$invalidMalformed.record.log, [string]$invalidUnknown.record.log, [string]$invalidDuplicate.record.log, [string]$invalidMissing.record.log, [string]$targetedTests.log)

Add-BooleanFeatureAssessment -Ctx $ctx -Id 'excellent.why_not_matched' -Level 'excellent' -Category 'explain' -Requirement 'Why policy did not match' -Ok $whyNotStructured -Evidence @([string]$coreExplainJSON.log)
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'excellent.window_conditions' -Level 'excellent' -Category 'algorithm' -Requirement 'Window conditions count within time' -Ok $windowOk -Evidence @([string]$windowRun.log, (Get-RelativeResultPath -Ctx $ctx -Path $windowOut))
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'excellent.html_report' -Level 'excellent' -Category 'format' -Requirement 'HTML report' -Ok ($htmlOk -and $reportSemanticsOk) -Evidence @((Get-RelativeResultPath -Ctx $ctx -Path $coreHTML), [string]$coreRun.log)
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'excellent.large_benchmark' -Level 'excellent' -Category 'performance' -Requirement 'Benchmark 100000 events and 100 policies' -Ok ($benchOk -and $largeOk) -Evidence @([string]$benchRun.log, [string]$largeRun.log, 'outputs/large_metrics.json')

$testFiles = @(Get-ChildItem -LiteralPath $ctx.RepoRoot -Recurse -File -Filter '*_test.go' -ErrorAction SilentlyContinue | Where-Object { $_.FullName -notlike '*\.check-results\*' })
$testFunctions = @($testFiles | Select-String -Pattern '^\s*func\s+Test[A-Za-z0-9_]+\s*\(' -ErrorAction SilentlyContinue)
$benchFunctions = @($testFiles | Select-String -Pattern '^\s*func\s+Benchmark[A-Za-z0-9_]+\s*\(' -ErrorAction SilentlyContinue)
$readmePath = Join-Path $ctx.RepoRoot 'README.md'
$makefilePath = Join-Path $ctx.RepoRoot 'Makefile'
$solutionPath = Join-Path $ctx.RepoRoot 'docs\reshenie.md'
$makefileText = if (Test-Path -LiteralPath $makefilePath) { [string](Get-Content -LiteralPath $makefilePath -Raw) } else { '' }
$controlScenarioPath = Join-Path $ctx.RepoRoot 'testdata\control\scenario.yaml'
$controlPoliciesPath = Join-Path $ctx.RepoRoot 'testdata\control\policies.yaml'
$controlExpectedPath = Join-Path $ctx.RepoRoot 'testdata\control\expected_alerts.jsonl'
$controlScenarioLines = if (Test-Path -LiteralPath $controlScenarioPath) { Get-CheckLineCount -Path $controlScenarioPath } else { 0 }
$controlPoliciesLines = if (Test-Path -LiteralPath $controlPoliciesPath) { Get-CheckLineCount -Path $controlPoliciesPath } else { 0 }
$controlDataOk = (Test-Path -LiteralPath $controlExpectedPath) -and ($controlScenarioLines -ge 20 -and $controlScenarioLines -le 50) -and ($controlPoliciesLines -ge 20 -and $controlPoliciesLines -le 50)

Add-BooleanFeatureAssessment -Ctx $ctx -Id 'engineering.unit_tests_present' -Level 'engineering' -Category 'tests' -Requirement 'Go unit tests are present' -Ok ($testFunctions.Count -gt 0) -Evidence @([string]$goTestAll.log)
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'engineering.benchmarks_present' -Level 'engineering' -Category 'benchmarks' -Requirement 'Go benchmark tests are present' -Ok ($benchFunctions.Count -gt 0) -Evidence @([string]$benchRun.log)
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'engineering.go_test_passes' -Level 'engineering' -Category 'tests' -Requirement 'go test ./... passes' -Ok (Test-CommandSuccess -Record $goTestAll) -Evidence @([string]$goTestAll.log)
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'engineering.make_test_runs' -Level 'engineering' -Category 'reproducibility' -Requirement 'make test passes' -Ok ($makeTest -ne $null -and (Test-CommandSuccess -Record $makeTest)) -Evidence @(if ($makeTest -ne $null) { [string]$makeTest.log } else { 'commands.jsonl' })
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'engineering.make_bench_runs' -Level 'engineering' -Category 'reproducibility' -Requirement 'make bench passes' -Ok ($makeBench -ne $null -and (Test-CommandSuccess -Record $makeBench)) -Evidence @(if ($makeBench -ne $null) { [string]$makeBench.log } else { 'commands.jsonl' })
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'engineering.make_demo_runs' -Level 'engineering' -Category 'reproducibility' -Requirement 'make demo passes' -Ok ($makeDemo -ne $null -and (Test-CommandSuccess -Record $makeDemo)) -Evidence @(if ($makeDemo -ne $null) { [string]$makeDemo.log } else { 'commands.jsonl' })
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'engineering.readme' -Level 'engineering' -Category 'documentation' -Requirement 'README.md exists and is not empty' -Ok ((Test-Path -LiteralPath $readmePath) -and ((Get-Item -LiteralPath $readmePath).Length -gt 100)) -Evidence @('repo_snapshot/README.md')
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'engineering.make_test' -Level 'engineering' -Category 'reproducibility' -Requirement 'Makefile has target test' -Ok ($makefileText -match '(?m)^\s*test\s*:') -Evidence @('repo_snapshot/Makefile')
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'engineering.make_bench' -Level 'engineering' -Category 'reproducibility' -Requirement 'Makefile has target bench' -Ok ($makefileText -match '(?m)^\s*bench\s*:') -Evidence @('repo_snapshot/Makefile')
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'engineering.make_demo' -Level 'engineering' -Category 'reproducibility' -Requirement 'Makefile has target demo' -Ok ($makefileText -match '(?m)^\s*demo\s*:') -Evidence @('repo_snapshot/Makefile')
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'engineering.control_data' -Level 'engineering' -Category 'reproducibility' -Requirement 'Fixed testdata/control set exists' -Ok $controlDataOk -Evidence @('repo_snapshot/README.md', 'repo_snapshot/Makefile')
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'engineering.solution_doc' -Level 'engineering' -Category 'documentation' -Requirement 'Non-empty docs/reshenie.md exists' -Ok ((Test-Path -LiteralPath $solutionPath) -and ((Get-Item -LiteralPath $solutionPath).Length -gt 100)) -Evidence @('repo_snapshot/docs/reshenie.md')

Complete-Check -Ctx $ctx -Extra @{
    hidden_runner = 'Start-Process -WindowStyle Hidden'
    timeout_seconds = $script:CheckTimeoutSec
    gofmt_ok = (Test-CommandSuccess -Record $goFmt)
    core_alert_exact = $coreExact
    core_reasons_exact_order = $coreReasonsOk
    core_jsonl_exact = $coreJsonlOk
    any_matrix_ok = $anyOk
    exists_matrix_ok = $existsOk
    yaml_runtime_validation = $yamlRuntimeOk
    explain_structured_ok = $whyNotStructured
    window_runtime = $windowOk
    large_run_ok = $largeOk
    benchmark_ok = $benchOk
    large_peak_tree_working_set_bytes = $largeRun.peak_tree_working_set_bytes
    large_output_growth_checkpoints = $largeGrowthCheckpoints
    cleanup_ok = $cleanupOk
    cleanup_removed = @($cleanupRemoved)
    git_status_meta_path = 'meta/git_status_short.txt'
}
