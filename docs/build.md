# Build

FoldersGuard uses SQLCipher for project and share databases. SQLCipher is a CGO dependency, so release builds must be
built with CGO enabled and a working C compiler for the target platform.

## Windows AMD64

On Windows, install MSYS2 and the UCRT64 MinGW-w64 toolchain.

In an MSYS2 UCRT64 shell:

```text
pacman -Syu
pacman -S mingw-w64-ucrt-x86_64-gcc
```

From the repository root, build with the Windows build script:

```text
powershell -ExecutionPolicy Bypass -File .\scripts\build-windows-amd64.ps1
```

If MSYS2 is installed somewhere other than `C:\msys64`, pass the root path:

```text
powershell -ExecutionPolicy Bypass -File .\scripts\build-windows-amd64.ps1 -MSYS2Root D:\msys64
```

The script sets `CGO_ENABLED`, `CC`, `CXX`, and `PATH` for the current build process.

Manual PowerShell build:

```text
$env:Path = "C:\msys64\ucrt64\bin;$env:Path"
$env:CGO_ENABLED = "1"
$env:CC = "gcc"
$env:CXX = "g++"
wails build -platform windows/amd64
```

If the build reports that `gcc` is missing, reopen PowerShell after updating `PATH` or use the full compiler paths:

```text
$env:CGO_ENABLED = "1"
$env:CC = "C:\msys64\ucrt64\bin\gcc.exe"
$env:CXX = "C:\msys64\ucrt64\bin\g++.exe"
wails build -platform windows/amd64
```

## macOS Cross Build For Windows AMD64

Install MinGW-w64:

```text
brew install mingw-w64
```

Build with the Windows target C compiler:

```text
CGO_ENABLED=1 \
CC=x86_64-w64-mingw32-gcc \
CXX=x86_64-w64-mingw32-g++ \
wails build -platform windows/amd64
```

## Failure Modes

If CGO is disabled, SQLCipher cannot be built into the application. Build with `CGO_ENABLED=1`.

If the host C compiler is used for a different target platform, Go may fail while compiling `runtime/cgo`. Set `CC` and
`CXX` to compilers for the target platform.
