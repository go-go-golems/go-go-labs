#!/usr/bin/env python3

from __future__ import annotations

import argparse
import asyncio
import json
import os
from pathlib import Path

import zigpy.config as zconf
import zigpy_znp.api as znp_api
import zigpy_znp.commands as c
import zigpy_znp.config as znp_conf
import zigpy_znp.types as t
from zigpy_znp.commands.af import LatencyReq, TransmitOptions


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


def _default_device_from_db(db_path: Path) -> tuple[int | None, int | None]:
    if not db_path.exists():
        return None, None

    for line in db_path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line:
            continue
        try:
            obj = json.loads(line)
        except json.JSONDecodeError:
            continue

        if obj.get("type") in ("Router", "EndDevice") and isinstance(obj.get("nwkAddr"), int):
            eps = obj.get("epList")
            if isinstance(eps, list) and eps and isinstance(eps[0], int):
                return obj["nwkAddr"], eps[0]
            return obj["nwkAddr"], 1

    return None, None


def _parse_hex_bytes(value: str) -> bytes:
    v = value.strip().lower()
    if v.startswith("0x"):
        v = v[2:]
    v = v.replace(":", "").replace(" ", "")
    if len(v) % 2 != 0:
        raise ValueError("hex data must have an even number of nybbles")
    return bytes.fromhex(v)


def build_arg_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Send an AF.DataRequest containing a ZCL payload (raw or OnOff convenience)."
    )
    parser.add_argument("--port", default="/dev/ttyUSB0")
    parser.add_argument("--baudrate", type=int, default=115200)

    parser.add_argument(
        "--db",
        default=str(Path("data") / "database.db"),
        help="Zigbee2MQTT JSONL database (optional, used for defaults).",
    )
    parser.add_argument("--dst-nwk", type=_parse_u16, default=None)
    parser.add_argument("--dst-ep", type=_parse_u8, default=None)

    parser.add_argument("--src-ep", type=_parse_u8, default=20)
    parser.add_argument("--profile", type=_parse_u16, default=0x0104)
    parser.add_argument("--device-id", type=_parse_u16, default=0x0000)

    parser.add_argument("--cluster", type=_parse_u16, default=0x0006)
    parser.add_argument(
        "--data-hex",
        default=None,
        help="Raw ZCL payload (hex), e.g. 0x110101 for On/Off (frame_control=0x11, seq=0x01, cmd=0x01).",
    )
    parser.add_argument(
        "--onoff",
        choices=("on", "off", "toggle"),
        default=None,
        help="Convenience: build a minimal ZCL On/Off command.",
    )
    parser.add_argument(
        "--zcl-seq",
        type=_parse_u8,
        default=None,
        help="ZCL transaction sequence number (default: random-ish).",
    )
    parser.add_argument(
        "--disable-default-response",
        action="store_true",
        default=True,
        help="Set ZCL frame control DisableDefaultResponse=1 (default true).",
    )
    parser.add_argument(
        "--no-disable-default-response",
        dest="disable_default_response",
        action="store_false",
        help="Clear ZCL DisableDefaultResponse bit.",
    )
    parser.add_argument("--radius", type=_parse_u8, default=30)
    parser.add_argument(
        "--security",
        action="store_true",
        default=True,
        help="Set AF option ENABLE_SECURITY (default true).",
    )
    parser.add_argument(
        "--no-security",
        dest="security",
        action="store_false",
        help="Do not set AF option ENABLE_SECURITY.",
    )
    parser.add_argument(
        "--ack",
        action="store_true",
        default=True,
        help="Set AF option ACK_REQUEST (default true).",
    )
    parser.add_argument(
        "--no-ack",
        dest="ack",
        action="store_false",
        help="Do not request APS ACK.",
    )
    parser.add_argument(
        "--pin-toggle-skip-bootloader",
        action="store_true",
        help="Enable zigpy-znp RTS/DTR toggling to skip bootloader/reset (off by default).",
    )
    return parser


def build_zcl_onoff_payload(*, cmd: str, zcl_seq: int, disable_default_response: bool) -> bytes:
    cmd_id = {"off": 0x00, "on": 0x01, "toggle": 0x02}[cmd]
    frame_control = 0x01  # cluster specific
    if disable_default_response:
        frame_control |= 0x10
    return bytes([frame_control, zcl_seq, cmd_id])


async def run(args: argparse.Namespace) -> int:
    dst_nwk = args.dst_nwk
    dst_ep = args.dst_ep
    if dst_nwk is None or dst_ep is None:
        inferred_nwk, inferred_ep = _default_device_from_db(Path(args.db))
        dst_nwk = dst_nwk if dst_nwk is not None else inferred_nwk
        dst_ep = dst_ep if dst_ep is not None else inferred_ep

    if dst_nwk is None or dst_ep is None:
        raise SystemExit("Missing --dst-nwk/--dst-ep (and could not infer from --db)")

    if args.data_hex is None and args.onoff is None:
        raise SystemExit("Provide --data-hex or --onoff")
    if args.data_hex is not None and args.onoff is not None:
        raise SystemExit("Use only one of --data-hex or --onoff")

    zcl_seq = args.zcl_seq
    if zcl_seq is None:
        zcl_seq = int.from_bytes(os.urandom(1), "little")

    if args.onoff is not None:
        payload = build_zcl_onoff_payload(
            cmd=args.onoff,
            zcl_seq=zcl_seq,
            disable_default_response=args.disable_default_response,
        )
    else:
        payload = _parse_hex_bytes(args.data_hex)

    cfg = znp_conf.CONFIG_SCHEMA(
        {
            zconf.CONF_DEVICE: {
                zconf.CONF_DEVICE_PATH: args.port,
                zconf.CONF_DEVICE_BAUDRATE: args.baudrate,
                zconf.CONF_DEVICE_FLOW_CONTROL: None,
            }
            ,
            znp_conf.CONF_ZNP_CONFIG: {
                znp_conf.CONF_SKIP_BOOTLOADER: bool(args.pin_toggle_skip_bootloader),
            },
        }
    )

    znp = znp_api.ZNP(cfg)
    try:
        await znp.connect()
        await znp.request(c.ZDO.StartupFromApp.Req(StartDelay=0))

        # Register a client-style HA endpoint so we can send/receive ZCL.
        out_clusters = [t.ClusterId(args.cluster)]
        await znp.request(
            c.AF.Register.Req(
                Endpoint=t.uint8_t(args.src_ep),
                ProfileId=t.uint16_t(args.profile),
                DeviceId=t.uint16_t(args.device_id),
                DeviceVersion=t.uint8_t(0),
                LatencyReq=LatencyReq.NoLatencyReqs,
                InputClusters=t.ClusterIdList([]),
                OutputClusters=t.ClusterIdList(out_clusters),
            )
        )

        tsn = int.from_bytes(os.urandom(1), "little")
        opts = TransmitOptions.NONE
        if args.ack:
            opts |= TransmitOptions.ACK_REQUEST
        if args.security:
            opts |= TransmitOptions.ENABLE_SECURITY

        await znp.request(
            c.AF.DataRequest.Req(
                DstAddr=t.NWK(dst_nwk),
                DstEndpoint=t.uint8_t(dst_ep),
                SrcEndpoint=t.uint8_t(args.src_ep),
                ClusterId=t.ClusterId(args.cluster),
                TSN=t.uint8_t(tsn),
                Options=opts,
                Radius=t.uint8_t(args.radius),
                Data=t.ShortBytes(payload),
            )
        )

        confirm = await znp.wait_for_response(
            c.AF.DataConfirm.Callback(
                partial=True, Endpoint=t.uint8_t(args.src_ep), TSN=t.uint8_t(tsn)
            )
        )

        print(f"dst_nwk: {_hex_u16(dst_nwk)}")
        print(f"dst_ep: {dst_ep}")
        print(f"src_ep: {args.src_ep}")
        print(f"cluster: {_hex_u16(args.cluster)}")
        print(f"zcl_payload_hex: {payload.hex()}")
        print(f"af_tsn: 0x{tsn:02x}")
        print(f"data_confirm_status: {confirm.Status!s}")

        return 0
    finally:
        await znp.disconnect()


def main() -> int:
    parser = build_arg_parser()
    args = parser.parse_args()
    return asyncio.run(run(args))


if __name__ == "__main__":
    raise SystemExit(main())
