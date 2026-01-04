#!/usr/bin/env python3

import importlib.metadata as metadata
import sys


def pkg_version(name: str) -> str:
    try:
        return metadata.version(name)
    except metadata.PackageNotFoundError:
        return "<not installed>"


def main() -> None:
    print(f"python: {sys.executable}")
    print(f"python_version: {sys.version.split()[0]}")
    print(f"zigpy: {pkg_version('zigpy')}")
    print(f"zigpy-znp: {pkg_version('zigpy-znp')}")
    print(f"pyserial: {pkg_version('pyserial')}")


if __name__ == "__main__":
    main()

