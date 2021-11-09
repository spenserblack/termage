if (Test-Path -Path "$Env:ProgramFiles\termage") {
    Write-Host "Overwriting previous install of termage..."
} else {
    Write-Host "Installing termage..."
    Write-Host "Creating directory in Program Files..."
    New-Item -Path "$Env:ProgramFiles" -Name "termage" -ItemType "directory"
    Write-Host "Adding directory to path..."
    $path = [Environment]::GetEnvironmentVariable('Path', 'User')
    $newpath = $path + ";$Env:ProgramFiles\termage"
    [Environment]::SetEnvironmentVariable('Path', $newpath, 'User')
    Write-Host "Updated Path!"
    Write-Host "You may need to restart your PowerShell instance for this to take effect."
}
Invoke-WebRequest -OutFile "$Env:ProgramFiles\termage\termage.exe" "https://github.com/spenserblack/termage/releases/latest/download/termage-windows.exe"
