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

## macOS Release Signing And Notarization

macOS release distribution requires an Apple Developer account, a `Developer ID Application` certificate installed in
the login keychain, and notary service credentials.

The automated release script builds the Wails app, signs the `.app` bundle with the hardened runtime, creates a ZIP
archive for notarization, submits it to Apple, staples the notary ticket, and recreates the final ZIP artifact.

Local release settings should be stored in `config/macos-release.env`. The `config/` directory is ignored by git, and
`config.example/` contains templates that are safe to commit:

```text
mkdir -p config
cp config.example/macos-release.env config/macos-release.env
```

Edit `config/macos-release.env`:

```text
APPLE_SIGN_IDENTITY="Developer ID Application: Your Name (TEAMID)"
APPLE_NOTARY_KEYCHAIN_PROFILE="apple-notary"
WAILS_PLATFORM="darwin/universal"
DIST_DIR="dist/macos"
```

Then run:

```text
./scripts/build-macos-release.sh
```

The same workflow is available through Make:

```text
make macos-release
```

To store notary credentials in the macOS keychain:

```text
xcrun notarytool store-credentials apple-notary \
  --apple-id "apple-id@example.com" \
  --team-id "TEAMID" \
  --password "app-specific-password"
```

Alternatively, pass credentials through environment variables:

```text
APPLE_ID="apple-id@example.com" \
APPLE_TEAM_ID="TEAMID" \
APPLE_APP_PASSWORD="app-specific-password" \
./scripts/build-macos-release.sh
```

Useful script options:

- `WAILS_PLATFORM`: Wails platform target. Defaults to `darwin/universal`.
- `APPLE_SIGN_IDENTITY`: Developer ID Application signing identity. If omitted, the script uses the first installed
  `Developer ID Application` identity.
- `APPLE_NOTARY_KEYCHAIN_PROFILE`: keychain profile created by `xcrun notarytool store-credentials`.
- `DIST_DIR`: output directory. Defaults to `dist/macos`.
- `SKIP_BUILD=1`: sign and notarize an existing app bundle.
- `SKIP_NOTARIZE=1`: build and sign only, without submitting to Apple.
- `MACOS_RELEASE_ENV`: path to a different env file. Defaults to `config/macos-release.env`.

The default app bundle path is:

```text
build/bin/FoldersGuard.app
```

The final ZIP artifact is:

```text
dist/macos/FoldersGuard-macOS.zip
```

If the script cannot find a `Developer ID Application` identity, install the certificate from the Apple Developer
account or set `APPLE_SIGN_IDENTITY` explicitly. An `Apple Development` certificate is not sufficient for Developer ID
distribution and notarization.
