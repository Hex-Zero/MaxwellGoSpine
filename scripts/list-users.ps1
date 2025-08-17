param(
  [string] $BaseUrl = 'http://localhost:8080',
  [switch] $Wide,
  [switch] $Raw
)

# Fetch users
try {
  $resp = Invoke-RestMethod -Method GET -Uri "$BaseUrl/v1/users" -TimeoutSec 5
} catch {
  Write-Error "Failed to fetch users: $_"
  exit 1
}
$users = $resp.data
if(-not $users){ Write-Host 'No users found.'; exit 0 }

if($Raw){ $users | ConvertTo-Json -Depth 4; exit 0 }

# Column widths (dynamic if wide)
if($Wide){
  $fmt = '{0}  {1,-20} {2}'
} else {
  $fmt = '{0}  {1,-12} {2}'
}
Write-Host ("ID{0}NAME{1}EMAIL" -f (' ' * 34), (' ' * (if($Wide){8}else{0}))) -ForegroundColor Cyan
Write-Host ('-' * 90) -ForegroundColor DarkCyan
foreach($u in $users){
  $line = $fmt -f $u.id, $u.name, $u.email
  Write-Host $line
}

Write-Host "\nTotal: $($users.Count)  Page: $($resp.page)  PageSize: $($resp.page_size)" -ForegroundColor Cyan
