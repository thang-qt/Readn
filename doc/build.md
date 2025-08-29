## Compilation

Prerequisies:

* Go >= 1.23
* C Compiler (GCC / Clang / ...)
* Zig >= 0.14.0 (optional, for cross-compiling CLI versions)
* binutils (optional, for building Windows GUI version)

Get the source code:

    git clone https://github.com/thang-qt/Readn.git

Compile:

    # create cli for the host OS/architecture
    make host               # out/readn

    # create GUI, works only in the target OS
    make windows_amd64_gui  # out/windows_amd64_gui/readn.exe
    make windows_arm64_gui  # out/windows_arm64_gui/readn.exe
    make darwin_arm64_gui   # out/darwin_arm64_gui/readn.app
    make darwin_amd64_gui   # out/darwin_amd64_gui/readn.app

    # create cli, cross-compiles within any OS/architecture
    make linux_amd64
    make linux_arm64
    make linux_armv7
    make windows_amd64
    make windows_arm64

    # ... or build a docker image
    docker build -t readn -f etc/dockerfile .

## ARM compilation

The instructions below are to cross-compile *Readn* to `Linux/ARM*`.

Build:

    docker build -t readn.arm -f etc/dockerfile.arm .

Test:

    # inside host
    docker run -it --rm readn.arm

    # then, inside container
    cd /root/out
    qemu-aarch64 -L /usr/aarch64-linux-gnu/ readn.arm64

Extract files from images:

    CID=$(docker create readn.arm)
    docker cp -a "$CID:/root/out" .
    docker rm "$CID"
