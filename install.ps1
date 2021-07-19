$tempdir = [System.IO.Path]::GetTempPath()
$repo = [System.IO.Path]::GetRandomFileName()
$temprepo = Join-Path $tempdir $repo
git clone https://github.com/spenserblack/termage.git $temprepo
Push-Location $temprepo
$version = git describe --tags
go install -ldflags "-X main.version=${version}" .
Pop-Location
echo Removing $temprepo
Remove-Item -Recurse -Force $temprepo
