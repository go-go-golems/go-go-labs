#!/usr/bin/env python3

import inspect

import zigpy_znp.api as api


def print_section(title: str) -> None:
    print()
    print("=" * len(title))
    print(title)
    print("=" * len(title))


def main() -> None:
    print_section("ZNP methods containing 'connect'")
    print([name for name in dir(api.ZNP) if "connect" in name.lower()])

    for name in ["connect", "request", "frame_received", "wait_for_responses"]:
        print_section(f"ZNP.{name}")
        fn = getattr(api.ZNP, name)
        print(inspect.getsource(fn))


if __name__ == "__main__":
    main()

