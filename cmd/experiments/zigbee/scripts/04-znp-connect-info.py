#!/usr/bin/env python3

from __future__ import annotations

import argparse
import asyncio
import json
from pathlib import Path

import zigpy.config as zconf
import zigpy_znp.api as znp_api
import zigpy_znp.commands as c
import zigpy_znp.config as znp_conf
import zigpy_znp.types as t
from zigpy_znp.types.nvids import OsalNvIds


def _hex_u16(value: int) -> str:
    return f"0x{value:04x}"


def _hex_eui64(value: t.EUI64) -> str:
    # t.EUI64 renders nicely, but we want stable output
    return "0x" + bytes(value).hex()


def _default_from_db(db_path: Path) -> tuple[str | None, int | None]:
    if not db_path.exists():
        return None, None

    ieee = None
    nwk = None

    for line in db_path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line:
            continue
        try:
            obj = json.loads(line)
        except json.JSONDecodeError:
            continue

        if obj.get("type") in ("Router", "EndDevice"):
            ieee = obj.get("ieeeAddr")
            nwk = obj.get("nwkAddr")
            if isinstance(ieee, str) and isinstance(nwk, int):
                return ieee, nwk

    return None, None


def build_arg_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Connect to a TI ZNP coordinator and print basic SYS/NVRAM info."
    )
    parser.add_argument("--port", default="/dev/ttyUSB0")
    parser.add_argument("--baudrate", type=int, default=115200)
    parser.add_argument(
        "--flow-control",
        default=None,
        choices=("hardware", "software", "none", "None", "null"),
        help="Serial flow control (default: none).",
    )
    parser.add_argument(
        "--db",
        default=str(Path("data") / "database.db"),
        help="Zigbee2MQTT JSONL database (optional, used only for printing a known device).",
    )
    parser.add_argument(
        "--show-keys",
        action="store_true",
        help="Print network keys (sensitive).",
    )
    parser.add_argument(
        "--pin-toggle-skip-bootloader",
        action="store_true",
        help="Enable zigpy-znp RTS/DTR toggling to skip bootloader/reset (off by default).",
    )
    return parser


def _parse_flow_control(value: str | None) -> str | None:
    if value is None:
        return None
    if value in ("none", "None", "null"):
        return None
    return value


async def run(args: argparse.Namespace) -> int:
    flow_control = _parse_flow_control(args.flow_control)

    cfg = znp_conf.CONFIG_SCHEMA(
        {
            zconf.CONF_DEVICE: {
                zconf.CONF_DEVICE_PATH: args.port,
                zconf.CONF_DEVICE_BAUDRATE: args.baudrate,
                zconf.CONF_DEVICE_FLOW_CONTROL: flow_control,
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

        ping = await znp.request(c.SYS.Ping.Req())
        version = await znp.request(c.SYS.Version.Req())

        ieee = await znp.nvram.osal_read(OsalNvIds.EXTADDR, item_type=t.EUI64)
        nib = await znp.nvram.osal_read(OsalNvIds.NIB, item_type=t.NIB)

        print(f"port: {args.port}")
        print(f"capabilities: 0x{int(ping.Capabilities):08x}")
        print(f"zstack_feature_detected: {znp.version}")
        print(f"sys.version: {version.as_dict()}")
        print(f"coordinator_ieee: {_hex_eui64(ieee)}")
        print(f"nwk_addr: {_hex_u16(int(nib.nwkDevAddress))}")
        print(f"pan_id: {_hex_u16(int(nib.nwkPanId))}")
        print(f"ext_pan_id: {_hex_eui64(nib.extendedPANID)}")
        print(f"channel: {int(nib.nwkLogicalChannel)}")
        print(f"security_level: {int(nib.SecurityLevel)}")
        print(f"nwk_key_loaded: {bool(nib.nwkKeyLoaded)}")

        if args.show_keys:
            key_desc = await znp.nvram.osal_read(
                OsalNvIds.NWK_ACTIVE_KEY_INFO, item_type=t.NwkKeyDesc
            )
            print(f"nwk_key_seq: {int(key_desc.KeySeqNum)}")
            print(f"nwk_key: {bytes(key_desc.Key).hex()}")

        db_ieee, db_nwk = _default_from_db(Path(args.db))
        if db_ieee is not None and db_nwk is not None:
            print(f"example_device_from_db_ieee: {db_ieee}")
            print(f"example_device_from_db_nwk: {_hex_u16(db_nwk)}")

        return 0
    finally:
        await znp.disconnect()


def main() -> int:
    parser = build_arg_parser()
    args = parser.parse_args()
    return asyncio.run(run(args))


if __name__ == "__main__":
    raise SystemExit(main())
