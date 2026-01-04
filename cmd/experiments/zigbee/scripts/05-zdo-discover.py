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


def _parse_u16(value: str) -> int:
    value = value.strip().lower()
    if value.startswith("0x"):
        return int(value, 16)
    return int(value, 10)


def _hex_u16(value: int) -> str:
    return f"0x{value:04x}"


def _default_nwk_from_db(db_path: Path) -> int | None:
    if not db_path.exists():
        return None

    for line in db_path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line:
            continue
        try:
            obj = json.loads(line)
        except json.JSONDecodeError:
            continue

        if obj.get("type") in ("Router", "EndDevice") and isinstance(obj.get("nwkAddr"), int):
            return obj["nwkAddr"]

    return None


def build_arg_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="ZDO discovery: Active Endpoints + Simple Descriptor per endpoint."
    )
    parser.add_argument("--port", default="/dev/ttyUSB0")
    parser.add_argument("--baudrate", type=int, default=115200)
    parser.add_argument(
        "--nwk",
        type=_parse_u16,
        default=None,
        help="Target NWK address (e.g. 0x0038). If omitted, tries data/database.db.",
    )
    parser.add_argument(
        "--db",
        default=str(Path("data") / "database.db"),
        help="Zigbee2MQTT JSONL database (optional).",
    )
    return parser


async def run(args: argparse.Namespace) -> int:
    nwk = args.nwk
    if nwk is None:
        nwk = _default_nwk_from_db(Path(args.db))
    if nwk is None:
        raise SystemExit("Missing --nwk and could not infer from --db")

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

        # Active Endpoints
        await znp.request(
            c.ZDO.ActiveEpReq.Req(DstAddr=nwk, NWKAddrOfInterest=nwk)
        )
        active_rsp = await znp.wait_for_response(
            c.ZDO.ActiveEpRsp.Callback(partial=True, Src=nwk, NWK=nwk)
        )

        eps = list(active_rsp.ActiveEndpoints)
        print(f"nwk: {_hex_u16(nwk)}")
        print(f"active_endpoints: {eps}")

        # Simple Descriptor per EP
        for ep in eps:
            await znp.request(
                c.ZDO.SimpleDescReq.Req(
                    DstAddr=nwk, NWKAddrOfInterest=nwk, Endpoint=ep
                )
            )
            simple_rsp = await znp.wait_for_response(
                c.ZDO.SimpleDescRsp.Callback(partial=True, Src=nwk, NWK=nwk)
            )

            desc = simple_rsp.SimpleDescriptor
            in_clusters = [int(x) for x in desc.InputClusters]
            out_clusters = [int(x) for x in desc.OutputClusters]

            print()
            print(f"endpoint: {ep}")
            print(f"  profile_id: {_hex_u16(int(desc.ProfileId))}")
            print(f"  device_id: {_hex_u16(int(desc.DeviceId))}")
            print(f"  device_version: {int(desc.DeviceVersion)}")
            print(f"  in_clusters: {[ _hex_u16(x) for x in in_clusters ]}")
            print(f"  out_clusters: {[ _hex_u16(x) for x in out_clusters ]}")

        return 0
    finally:
        await znp.disconnect()


def main() -> int:
    parser = build_arg_parser()
    args = parser.parse_args()
    return asyncio.run(run(args))


if __name__ == "__main__":
    raise SystemExit(main())
