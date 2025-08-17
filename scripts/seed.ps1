param(
  [string[]] $Users = @('Alice:alice@example.com','Bob:bob@example.com','Carol:carol@example.com','Dave:dave@example.com','Eve:eve@example.com'),
  [string] $BaseUrl = 'http://localhost:8080'
)

function Wait-Health {
  param([string]$Url,[int]$TimeoutSec=20)
  $stop = (Get-Date).AddSeconds($TimeoutSec)
  while((Get-Date) -lt $stop){
    try { $r = Invoke-RestMethod -Method GET -Uri "$Url/healthz" -TimeoutSec 2; if($r.status -eq 'ok'){ return $true } } catch {}
    Start-Sleep -Milliseconds 500
  }
  return $false
}

if(-not (Wait-Health -Url $BaseUrl)){ Write-Error "API not healthy at $BaseUrl"; exit 1 }

$created = @()
foreach($u in $Users){
  $parts = $u.Split(':',2)
  if($parts.Count -ne 2){ Write-Warning "Skip malformed entry $u"; continue }
  $name = $parts[0]; $email = $parts[1]
  $payload = @{ name = $name; email = $email } | ConvertTo-Json -Compress
  try {
    $resp = Invoke-RestMethod -Method POST -Uri "$BaseUrl/v1/users" -ContentType 'application/json' -Body $payload -TimeoutSec 5
    $created += [pscustomobject]@{ Name=$resp.name; Email=$resp.email; ID=$resp.id }
    Write-Host "Created user $name ($email) -> $($resp.id)"
  } catch {
    Write-Warning "Failed to create $name <$email>: $_"
  }
}

Write-Host "\nSummary:" -ForegroundColor Cyan
$created | Format-Table -AutoSize
