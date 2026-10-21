# Input: $Action (start|stop|status), $Ssid, $Pass. Output: one JSON line.
$ErrorActionPreference = 'Stop'
function Out($h) { $h | ConvertTo-Json -Compress }
try {
  Add-Type -AssemblyName System.Runtime.WindowsRuntime
  $asTask = ([System.WindowsRuntimeSystemExtensions].GetMethods() | Where-Object {
    $_.Name -eq 'AsTask' -and $_.GetParameters().Count -eq 1 -and
    $_.GetParameters()[0].ParameterType.Name -eq 'IAsyncOperation`1' })[0]
  $asTaskAction = ([System.WindowsRuntimeSystemExtensions].GetMethods() | Where-Object {
    $_.Name -eq 'AsTask' -and $_.GetParameters().Count -eq 1 -and
    $_.GetParameters()[0].ParameterType.Name -eq 'IAsyncAction' })[0]
  function Await($op, [Type]$t) { $task = $asTask.MakeGenericMethod($t).Invoke($null, @($op)); $task.Wait(-1) | Out-Null; $task.Result }
  function AwaitAction($op) { $task = $asTaskAction.Invoke($null, @($op)); $task.Wait(-1) | Out-Null }

  [Windows.Networking.Connectivity.NetworkInformation,Windows.Networking.Connectivity,ContentType=WindowsRuntime] | Out-Null
  [Windows.Networking.NetworkOperators.NetworkOperatorTetheringManager,Windows.Networking.NetworkOperators,ContentType=WindowsRuntime] | Out-Null
  [Windows.Networking.NetworkOperators.NetworkOperatorTetheringAccessPointConfiguration,Windows.Networking.NetworkOperators,ContentType=WindowsRuntime] | Out-Null

  # Share the tunnel card first, then whatever has internet now.
  $profiles = [Windows.Networking.Connectivity.NetworkInformation]::GetConnectionProfiles()
  $cands = @()
  $cands += $profiles | Where-Object { $_.ProfileName -eq 'TelTun' }
  $inet = [Windows.Networking.Connectivity.NetworkInformation]::GetInternetConnectionProfile()
  if ($inet) { $cands += $inet }

  $tm = $null; $prof = ''; $lastErr = 'no internet connection to share'
  foreach ($p in $cands) {
    try {
      $cap = [Windows.Networking.NetworkOperators.NetworkOperatorTetheringManager]::GetTetheringCapabilityFromConnectionProfile($p)
      if ("$cap" -ne 'Enabled') { $lastErr = "tethering: $cap"; continue }
      $tm = [Windows.Networking.NetworkOperators.NetworkOperatorTetheringManager]::CreateFromConnectionProfile($p)
      $prof = $p.ProfileName
      break
    } catch { $lastErr = $_.Exception.Message }
  }
  if (-not $tm) { Out @{ ok = $false; state = 'Unknown'; msg = $lastErr }; exit }

  if ($Action -eq 'start') {
    if ($Ssid -or $Pass) {
      $c = $tm.GetCurrentAccessPointConfiguration()
      if ($Ssid) { $c.Ssid = $Ssid }
      if ($Pass) { $c.Passphrase = $Pass }
      AwaitAction ($tm.ConfigureAccessPointAsync($c))
    }
    if ("$($tm.TetheringOperationalState)" -ne 'On') {
      $r = Await ($tm.StartTetheringAsync()) ([Windows.Networking.NetworkOperators.NetworkOperatorTetheringOperationResult])
      if ("$($r.Status)" -ne 'Success') { Out @{ ok = $false; state = "$($tm.TetheringOperationalState)"; msg = "$($r.Status) $($r.AdditionalErrorMessage)" }; exit }
    }
  } elseif ($Action -eq 'stop') {
    if ("$($tm.TetheringOperationalState)" -ne 'Off') {
      $r = Await ($tm.StopTetheringAsync()) ([Windows.Networking.NetworkOperators.NetworkOperatorTetheringOperationResult])
      if ("$($r.Status)" -ne 'Success') { Out @{ ok = $false; state = "$($tm.TetheringOperationalState)"; msg = "$($r.Status) $($r.AdditionalErrorMessage)" }; exit }
    }
  }

  $c = $tm.GetCurrentAccessPointConfiguration()
  Out @{ ok = $true; state = "$($tm.TetheringOperationalState)"; ssid = $c.Ssid; pass = $c.Passphrase;
         clients = [int]$tm.ClientCount; max = [int]$tm.MaxClientCount; profile = $prof }
} catch {
  Out @{ ok = $false; state = 'Unknown'; msg = $_.Exception.Message }
}
