#!/usr/bin/env python3

from zigpy_znp.frames import TransportFrame
import zigpy_znp.commands as c


def main() -> None:
    req = c.SYS.Version.Req()
    general = req.to_frame()
    transport = TransportFrame(general)

    print(f"request: {req!r}")
    print(f"general_frame: {general!r}")
    raw = transport.serialize()
    print(f"transport_hex: {raw.hex()}")


if __name__ == "__main__":
    main()

