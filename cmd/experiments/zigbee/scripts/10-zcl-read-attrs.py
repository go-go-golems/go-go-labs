#!/usr/bin/env python3

from __future__ import annotations

import argparse
import asyncio
import os
from dataclasses import dataclass

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


def build_arg_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Send ZCL Read Attributes and print the Read Attributes Response."
    )
    parser.add_argument("--port", default="/dev/ttyUSB0")
    parser.add_argument("--baudrate", type=int, default=115200)

    parser.add_argument("--dst-nwk", type=_parse_u16, required=True)
    parser.add_argument("--dst-ep", type=_parse_u8, required=True)
    parser.add_argument("--src-ep", type=_parse_u8, default=20)

    parser.add_argument("--cluster", type=_parse_u16, required=True)
    parser.add_argument(
        "--attr",
        type=_parse_u16,
        action="append",
        required=True,
        help="Attribute ID to read (repeatable).",
    )

    parser.add_argument(
        "--zcl-seq",
        type=_parse_u8,
        default=None,
        help="ZCL transaction sequence number (default: random).",
    )
    parser.add_argument(
        "--disable-default-response",
        action="store_true",
        default=True,
        help="Set ZCL DisableDefaultResponse=1 (default true).",
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
        "--timeout",
        type=float,
        default=10.0,
        help="Seconds to wait for Read Attributes Response (default: 10).",
    )
    parser.add_argument(
        "--pin-toggle-skip-bootloader",
        action="store_true",
        help="Enable zigpy-znp RTS/DTR toggling to skip bootloader/reset (off by default).",
    )
    return parser


def build_zcl_read_attributes_payload(
    *, zcl_seq: int, disable_default_response: bool, attr_ids: list[int]
) -> bytes:
    # Frame Control:
    #   frame_type=0 (foundation)
    #   manufacturer_specific=0
    #   direction=0 (client->server)
    #   disable_default_response=bit4
    frame_control = 0x00
    if disable_default_response:
        frame_control |= 0x10
    command_id = 0x00  # Read Attributes

    payload = bytearray([frame_control, zcl_seq, command_id])
    for attr in attr_ids:
        payload += int(attr).to_bytes(2, "little")
    return bytes(payload)


def _zcl_skip_header(payload: bytes) -> tuple[int, int, int, bytes]:
    if len(payload) < 3:
        raise ValueError("payload too short for ZCL header")
    fc = payload[0]
    manuf_specific = (fc >> 2) & 0x01
    idx = 1
    if manuf_specific:
        if len(payload) < 5:
            raise ValueError("payload too short for manufacturer-specific ZCL header")
        idx += 2
    tsn = payload[idx]
    cmd = payload[idx + 1]
    body = payload[idx + 2 :]
    return fc, tsn, cmd, body


@dataclass
class ReadAttrRecord:
    attr_id: int
    status: int
    data_type: int | None
    raw_value: bytes | None


def _datatype_len(data_type: int) -> int | None:
    # Minimal subset: enough for common plug attributes
    fixed = {
        0x10: 1,  # bool
        0x18: 1,  # bitmap8
        0x20: 1,  # uint8
        0x21: 2,  # uint16
        0x23: 4,  # uint32
        0x28: 1,  # int8
        0x29: 2,  # int16
        0x2B: 4,  # int32
        0x30: 1,  # enum8
        0x31: 2,  # enum16
    }
    return fixed.get(data_type)


def parse_read_attributes_response(body: bytes) -> list[ReadAttrRecord]:
    records: list[ReadAttrRecord] = []
    i = 0
    while i + 3 <= len(body):
        attr_id = int.from_bytes(body[i : i + 2], "little")
        status = body[i + 2]
        i += 3

        if status != 0x00:
            records.append(ReadAttrRecord(attr_id=attr_id, status=status, data_type=None, raw_value=None))
            continue

        if i >= len(body):
            break
        data_type = body[i]
        i += 1

        size = _datatype_len(data_type)
        if size is None or i + size > len(body):
            # Unknown or truncated type; store remaining as raw
            raw = body[i:]
            records.append(ReadAttrRecord(attr_id=attr_id, status=status, data_type=data_type, raw_value=raw))
            break

        raw = body[i : i + size]
        i += size
        records.append(ReadAttrRecord(attr_id=attr_id, status=status, data_type=data_type, raw_value=raw))

    return records


def decode_value(data_type: int, raw: bytes) -> int | bool | None:
    if data_type == 0x10:
        return bool(raw[0])
    if data_type in (0x18, 0x20, 0x28, 0x30):
        return int.from_bytes(raw, "little", signed=(data_type in (0x28,)))
    if data_type in (0x21, 0x31, 0x29):
        return int.from_bytes(raw, "little", signed=(data_type == 0x29))
    if data_type in (0x23, 0x2B):
        return int.from_bytes(raw, "little", signed=(data_type == 0x2B))
    return None


async def run(args: argparse.Namespace) -> int:
    zcl_seq = args.zcl_seq
    if zcl_seq is None:
        zcl_seq = int.from_bytes(os.urandom(1), "little")

    zcl = build_zcl_read_attributes_payload(
        zcl_seq=zcl_seq,
        disable_default_response=args.disable_default_response,
        attr_ids=args.attr,
    )

    cfg = znp_conf.CONFIG_SCHEMA(
        {
            zconf.CONF_DEVICE: {
                zconf.CONF_DEVICE_PATH: args.port,
                zconf.CONF_DEVICE_BAUDRATE: args.baudrate,
                zconf.CONF_DEVICE_FLOW_CONTROL: None,
            },
            znp_conf.CONF_ZNP_CONFIG: {
                znp_conf.CONF_SKIP_BOOTLOADER: bool(args.pin_toggle_skip_bootloader),
            },
        }
    )

    znp = znp_api.ZNP(cfg)
    try:
        await znp.connect()
        await znp.request(c.ZDO.StartupFromApp.Req(StartDelay=0))

        # Register endpoint so responses have somewhere to go.
        clusters = [t.ClusterId(args.cluster)]
        await znp.request(
            c.AF.Register.Req(
                Endpoint=t.uint8_t(args.src_ep),
                ProfileId=t.uint16_t(0x0104),
                DeviceId=t.uint16_t(0x0000),
                DeviceVersion=t.uint8_t(0),
                LatencyReq=LatencyReq.NoLatencyReqs,
                InputClusters=t.ClusterIdList(clusters),
                OutputClusters=t.ClusterIdList(clusters),
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
                DstAddr=t.NWK(args.dst_nwk),
                DstEndpoint=t.uint8_t(args.dst_ep),
                SrcEndpoint=t.uint8_t(args.src_ep),
                ClusterId=t.ClusterId(args.cluster),
                TSN=t.uint8_t(tsn),
                Options=opts,
                Radius=t.uint8_t(args.radius),
                Data=t.ShortBytes(zcl),
            )
        )

        confirm = await asyncio.wait_for(
            znp.wait_for_response(
                c.AF.DataConfirm.Callback(
                    partial=True, Endpoint=t.uint8_t(args.src_ep), TSN=t.uint8_t(tsn)
                )
            ),
            timeout=5.0,
        )

        print(f"dst_nwk: {_hex_u16(args.dst_nwk)} dst_ep: {args.dst_ep}")
        print(f"cluster: {_hex_u16(args.cluster)} attrs: {[ _hex_u16(a) for a in args.attr ]}")
        print(f"zcl_seq: 0x{zcl_seq:02x} zcl_hex: {zcl.hex()}")
        print(f"af_tsn: 0x{tsn:02x} data_confirm_status: {confirm.Status!s}")

        incoming = await asyncio.wait_for(
            znp.wait_for_response(
                c.AF.IncomingMsg.Callback(
                    partial=True,
                    SrcAddr=t.NWK(args.dst_nwk),
                    SrcEndpoint=t.uint8_t(args.dst_ep),
                    DstEndpoint=t.uint8_t(args.src_ep),
                    ClusterId=t.ClusterId(args.cluster),
                )
            ),
            timeout=args.timeout,
        )

        payload = bytes(incoming.Data)
        fc, resp_seq, cmd, body = _zcl_skip_header(payload)
        print(f"incoming_lqi: {int(incoming.LQI)}")
        print(f"incoming_zcl_fc: 0x{fc:02x} tsn: 0x{resp_seq:02x} cmd: 0x{cmd:02x}")
        print(f"incoming_zcl_hex: {payload.hex()}")

        if cmd == 0x01:  # Read Attributes Response
            records = parse_read_attributes_response(body)
            for r in records:
                if r.status != 0x00:
                    print(f"attr {_hex_u16(r.attr_id)} status=0x{r.status:02x}")
                    continue
                assert r.data_type is not None and r.raw_value is not None
                decoded = decode_value(r.data_type, r.raw_value)
                print(
                    f"attr {_hex_u16(r.attr_id)} type=0x{r.data_type:02x} raw={r.raw_value.hex()} decoded={decoded}"
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
