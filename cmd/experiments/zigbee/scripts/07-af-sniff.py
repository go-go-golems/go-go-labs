#!/usr/bin/env python3

from __future__ import annotations

import argparse
import asyncio
import time

import zigpy.config as zconf
import zigpy_znp.api as znp_api
import zigpy_znp.commands as c
import zigpy_znp.config as znp_conf
import zigpy_znp.types as t


def _parse_u8(value: str) -> int:
    value = value.strip().lower()
    if value.startswith("0x"):
        return int(value, 16) & 0xFF
    return int(value, 10) & 0xFF


def _parse_u16(value: str) -> int:
    value = value.strip().lower()
    if value.startswith("0x"):
        return int(value, 16) & 0xFFFF
    return int(value, 10) & 0xFFFF


def _hex_u16(value: int) -> str:
    return f"0x{value:04x}"


def decode_zcl_header(payload: bytes) -> dict[str, int] | None:
    if len(payload) < 3:
        return None
    fc = payload[0]
    frame_type = fc & 0x03
    manufacturer_specific = (fc >> 2) & 0x01
    direction = (fc >> 3) & 0x01
    disable_default_response = (fc >> 4) & 0x01
    idx = 1
    if manufacturer_specific:
        if len(payload) < 5:
            return None
        idx += 2
    if len(payload) < idx + 2:
        return None
    tsn = payload[idx]
    cmd_id = payload[idx + 1]
    return {
        "frame_control": fc,
        "frame_type": frame_type,
        "manufacturer_specific": manufacturer_specific,
        "direction": direction,
        "disable_default_response": disable_default_response,
        "zcl_tsn": tsn,
        "zcl_cmd": cmd_id,
    }


def build_arg_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Listen for AF.IncomingMsg and print frames.")
    parser.add_argument("--port", default="/dev/ttyUSB0")
    parser.add_argument("--baudrate", type=int, default=115200)
    parser.add_argument("--src-ep", type=_parse_u8, default=20)
    parser.add_argument("--profile", type=_parse_u16, default=0x0104)
    parser.add_argument(
        "--clusters",
        default="0x0006,0x0b04,0x0702",
        help="Comma-separated cluster IDs to register as client (out clusters).",
    )
    parser.add_argument("--seconds", type=int, default=60)
    return parser


async def run(args: argparse.Namespace) -> int:
    out_clusters = []
    for part in args.clusters.split(","):
        part = part.strip()
        if not part:
            continue
        out_clusters.append(t.ClusterId(_parse_u16(part)))

    cfg = znp_conf.CONFIG_SCHEMA(
        {
            zconf.CONF_DEVICE: {
                zconf.CONF_DEVICE_PATH: args.port,
                zconf.CONF_DEVICE_BAUDRATE: args.baudrate,
                zconf.CONF_DEVICE_FLOW_CONTROL: None,
            }
        }
    )

    znp = znp_api.ZNP(cfg)
    try:
        await znp.connect()

        await znp.request(
            c.AF.Register.Req(
                Endpoint=t.uint8_t(args.src_ep),
                ProfileId=t.uint16_t(args.profile),
                DeviceId=t.uint16_t(0x0000),
                DeviceVersion=t.uint8_t(0),
                LatencyReq=c.AF.LatencyReq.NoLatencyReqs,
                InputClusters=t.ClusterIdList([]),
                OutputClusters=t.ClusterIdList(out_clusters),
            )
        )

        print(f"listening_seconds: {args.seconds}")
        print(f"src_ep: {args.src_ep}")
        print(f"registered_out_clusters: {[ _hex_u16(int(x)) for x in out_clusters ]}")
        print()

        end = time.time() + args.seconds
        while time.time() < end:
            remaining = end - time.time()
            timeout = max(0.1, min(5.0, remaining))
            try:
                msg = await asyncio.wait_for(
                    znp.wait_for_response(c.AF.IncomingMsg.Callback(partial=True)),
                    timeout=timeout,
                )
            except TimeoutError:
                continue
            data = bytes(msg.Data)
            zcl = decode_zcl_header(data)

            print(
                "incoming:",
                f"src_nwk={_hex_u16(int(msg.SrcAddr))}",
                f"src_ep={int(msg.SrcEndpoint)}",
                f"dst_ep={int(msg.DstEndpoint)}",
                f"cluster={_hex_u16(int(msg.ClusterId))}",
                f"lqi={int(msg.LQI)}",
                f"tsn=0x{int(msg.TSN):02x}",
                f"payload_hex={data.hex()}",
            )
            if zcl is not None:
                print(
                    "  zcl:",
                    f"fc=0x{zcl['frame_control']:02x}",
                    f"type={zcl['frame_type']}",
                    f"dir={zcl['direction']}",
                    f"ddr={zcl['disable_default_response']}",
                    f"zcl_tsn=0x{zcl['zcl_tsn']:02x}",
                    f"cmd=0x{zcl['zcl_cmd']:02x}",
                )

        return 0
    finally:
        await znp.disconnect()


def main() -> int:
    parser = build_arg_parser()
    args = parser.parse_args()
    return asyncio.run(run(args))


if __name__ == "__main__":
    raise SystemExit(main())
