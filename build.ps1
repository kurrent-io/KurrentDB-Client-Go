[CmdletBinding()]
Param(
    [switch] $generateProtos
)

$ErrorActionPreference = "Stop"

function Exec {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true, Position = 0)][scriptblock]$Command,
        [Parameter(Mandatory = $false, Position = 1)][string]$ErrorMessage = ("Failed executing {0}" -F $Command)
    )

    . $Command
    if ($LASTEXITCODE -ne 0) {
        throw ("Exec: " + $ErrorMessage)
    }
}

$protobufVersion = "26.0"
$protocTarball = "protoc-$protobufVersion-win64.zip"
$protocUrl = "https://github.com/protocolbuffers/protobuf/releases/download/v$protobufVersion/$protocTarball"

go version

# Required tools
New-Item -Path . -Name "tools" -ItemType "directory" -Force | Out-Null
Push-Location tools

if (-not (Test-Path -LiteralPath "protobuf")) {
    Write-Host "Installing protoc $protobufVersion locally..."

    Invoke-WebRequest -Uri $protocUrl -OutFile $protocTarball
    Expand-Archive -LiteralPath $protocTarball -DestinationPath "protobuf" -Force
    Remove-Item $protocTarball

    Write-Host "done."
}

Pop-Location
# end

if ($generateProtos) {
    Exec { go install google.golang.org/protobuf/cmd/protoc-gen-go@latest } "Cannot run go command"
    Exec { go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest } "Cannot run go command"

    # We don't check the env variables because Github actions don't set $GOPATH.
    $gopath = Exec { go env GOPATH } "Cannot get $$GOPATH"

    Write-Host "Generating gRPC code..."

    $baseOutDir = "$pwd\protos"

    Get-ChildItem -Path "protos" -Recurse -Filter "*.proto" | ForEach-Object {
        $file = $_
        $protoName = [System.IO.Path]::GetFileNameWithoutExtension($file.Name)
        $protoDir = $file.DirectoryName
        $outputDir = Join-Path $protoDir $protoName

        # Create the output directory for this proto file
        New-Item -Path $outputDir -ItemType "directory" -Force | Out-Null

        Write-Host "Compiling $($file.Name) into $protoName folder..."

        # Generate Go files with a custom go_package option to control the output
        # First, generate into a temp location, then move to the correct subfolder
        $tempDir = Join-Path $env:TEMP "protogen_$protoName"
        New-Item -Path $tempDir -ItemType "directory" -Force | Out-Null

        try {
            Exec { tools\protobuf\bin\protoc --proto_path=$pwd\protos --go_out=$tempDir --go-grpc_out=$tempDir --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --plugin=protoc-gen-go=$gopath\bin\protoc-gen-go.exe --plugin=protoc-gen-go-grpc=$gopath\bin\protoc-gen-go-grpc.exe $file } "Cannot run protoc"

            # Move generated files from temp directory to the correct subfolder
            Get-ChildItem -Path $tempDir -Recurse -Filter "*.pb.go" | ForEach-Object {
                Move-Item $_.FullName $outputDir -Force
            }
        } finally {
            # Clean up temp directory
            Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
        }

        Write-Host "done."
    }

    Write-Host "Code generation completed."
}

Write-Host "Compiling project..."
go build -v .\kurrentdb .\samples
Write-Host "done."
