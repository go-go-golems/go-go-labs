#!/usr/bin/env python3

from __future__ import annotations

import argparse
import time

import serial


SOF = 0xFE


def fcs(data: bytes) -> int:
    checksum = 0
    for b in data:
        checksum ^= b
    return checksum & 0xFF


def read_exact(ser: serial.Serial, n: int, timeout_s: float) -> bytes:
    deadline = time.time() + timeout_s
    out = bytearray()
    while len(out) < n:
        if time.time() > deadline:
            raise TimeoutError(f"timeout reading {n} bytes (got {len(out)})")
        chunk = ser.read(n - len(out))
        if chunk:
            out += chunk
        else:
            time.sleep(0.01)
    return bytes(out)


def read_frame(ser: serial.Serial, timeout_s: float = 2.0) -> tuple[int, int, bytes]:
    # Scan to SOF
    deadline = time.time() + timeout_s
    while True:
        if time.time() > deadline:
            raise TimeoutError("timeout waiting for SOF")
        b = ser.read(1)
        if not b:
            time.sleep(0.01)
            continue
        if b[0] == SOF:
            break

    length = read_exact(ser, 1, timeout_s)[0]
    hdr = read_exact(ser, 2, timeout_s)
    payload = read_exact(ser, length, timeout_s) if length else b""
    chk = read_exact(ser, 1, timeout_s)[0]

    calc = fcs(bytes([length]) + hdr + payload)
    if chk != calc:
        raise ValueError(
            f"bad fcs: got 0x{chk:02x} expected 0x{calc:02x} (hdr={hdr.hex()} payload={payload.hex()})"
        )

    cmd0 = hdr[0]
    cmd1 = hdr[1]
    return cmd0, cmd1, payload


def build_sys_ping_req() -> bytes:
    # SYS.Ping is SREQ id=0x01 in SYS subsystem (0x01)
    # Command header: [cmd0, cmd1]
    #   cmd0 = (type<<5) | subsystem
    #     type(SREQ)=1 => 0x20, subsystem SYS=0x01 => 0x21
    #   cmd1 = command id = 0x01
    length = 0x00
    cmd0 = 0x21
    cmd1 = 0x01
    chk = fcs(bytes([length, cmd0, cmd1]))
    return bytes([SOF, length, cmd0, cmd1, chk])


def main() -> int:
    parser = argparse.ArgumentParser(description="Raw ZNP: send SYS.Ping and print the SRSP.")
    parser.add_argument("--port", default="/dev/ttyUSB0")
    parser.add_argument("--baudrate", type=int, default=115200)
    parser.add_argument("--timeout", type=float, default=2.0)
    args = parser.parse_args()

    req = build_sys_ping_req()
    print(f"tx_hex: {req.hex()}")

    with serial.Serial(args.port, args.baudrate, timeout=0) as ser:
        ser.reset_input_buffer()
        ser.write(req)
        ser.flush()

        cmd0, cmd1, payload = read_frame(ser, timeout_s=args.timeout)
        print(f"rx_cmd0: 0x{cmd0:02x}")
        print(f"rx_cmd1: 0x{cmd1:02x}")
        print(f"rx_payload_hex: {payload.hex()}")

        # SYS.Ping SRSP payload is MTCapabilities (u32, little-endian)
        if len(payload) in (2, 4):
            caps = int.from_bytes(payload, "little")
            width = len(payload) * 2
            print(f"capabilities_u{len(payload)*8}: 0x{caps:0{width}x}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
