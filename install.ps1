$ErrorActionPreference = "Stop"

$repo = "infraspecdev/mini-heroku"
$file = "mini-windows-amd64.exe"
$url = "https://github.com/$repo/releases/latest/download/$file"

Write-Host "Downloading mini CLI..."

$output = "$env:TEMP\mini.exe"

Invoke-WebRequest -Uri $url -OutFile $output

$installDir = "$env:USERPROFILE\mini"

if(!(Test-Path $installDir)){
    New-Item -ItemType Directory -Path $installDir | Out-Null
}

Move-Item $output "$installDir\mini.exe" -Force

Write-Host "Adding to PATH..."

$envPath = [Environment]::GetEnvironmentVariable("Path", "User")

if ($envPath -notlike "*$installDir"){
    [Environment]::SetEnvironmentVariable(
        "Path",
        "$envPath;$installDir",
        "User"
    )
}

Write-Host "mini CLI  installed!"
Write-Host "Restart terminal and run: mini version"
