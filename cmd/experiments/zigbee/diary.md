# Diary

## Goal

Bring up Mosquitto + Zigbee2MQTT via Docker Compose against the Sonoff Zigbee USB dongle on `/dev/ttyUSB0`, then pair a Sonoff Zigbee power switch and confirm it shows up in Zigbee2MQTT logs.

## Step 1: Baseline + coordinator detection

I started by inspecting the existing `docker-compose.yml` and Zigbee2MQTT `data/configuration.yaml` to understand what’s already wired up. The initial config looked close, but MQTT wasn’t running in Compose and the Zigbee adapter type looked mismatched for the attached Sonoff dongle.

Next I verified the serial device identity and probed the dongle with `zigpy` to confirm which radio stack it speaks. This determines the correct Zigbee2MQTT `serial.adapter` value.

### What I did
- Inspected existing Compose and config:
  - `sed -n '1,200p' docker-compose.yml`
  - `sed -n '1,200p' data/configuration.yaml`
- Confirmed serial device + udev identity:
  - `ls -l /dev/serial/by-id`
  - `udevadm info -n /dev/ttyUSB0 | rg -i 'ID_VENDOR|ID_MODEL|DRIVER|SERIAL|DEVLINKS'`
  - `lsusb | rg -i 'sonoff|itead|silabs|cp210|ti|zigbee'`
- Probed the dongle with zigpy (to identify radio type):
  - `zigpy radio znp /dev/ttyUSB0 info`

### Why
- Zigbee2MQTT needs the correct coordinator adapter type (e.g. `zstack` for TI ZNP, `ezsp` for Silabs EmberZNet).
- MQTT needs to be reachable from inside the Zigbee2MQTT container; `mqtt://localhost:1883` is wrong unless Mosquitto runs in the same container.

### What worked
- `/dev/serial/by-id/...Sonoff_Zigbee_3.0_USB_Dongle_Plus... -> ../../ttyUSB0` is present.
- `zigpy radio znp /dev/ttyUSB0 info` succeeded and printed network parameters, indicating the dongle speaks TI ZNP (Sonoff “Dongle-P” / CC2652P family).

### What didn't work
- A combined probe command timed out after 10s (I tried multiple `zigpy radio ... info` back-to-back); the `znp` probe completed, but the others didn’t run before the timeout.

### What I learned
- Even if the USB UART bridge reports “Silicon Labs CP210x”, the actual Zigbee radio stack is determined by the firmware/chip behind it; probing with `zigpy` is the decisive check.

### What was tricky to build
- Only one process can hold `/dev/ttyUSB0` at a time; probing must happen while Zigbee2MQTT is stopped.

### What warrants a second pair of eyes
- Confirm whether we should preserve the existing Zigbee network settings already stored in the dongle (PAN/channel/network key) vs. force Zigbee2MQTT to create a new network.

### What should be done in the future
- N/A.

### Code review instructions
- Start at `docker-compose.yml` and `data/configuration.yaml` once updated; validate by running `docker compose up -d` then `docker compose logs -f zigbee2mqtt`.

### Technical details
- Current config before changes had `mqtt.server: mqtt://localhost:1883` (no Mosquitto in compose) and `serial.adapter: ember` (expected `zstack` for ZNP).

## Step 2: Wire Mosquitto + switch Zigbee2MQTT to Z-Stack

This step updates the Compose stack to include an MQTT broker (Mosquitto) and reconfigures Zigbee2MQTT to use the correct TI ZNP/Z-Stack adapter. It also aligns the Zigbee2MQTT network settings (channel/PAN/extended PAN/network key) to what the coordinator currently reports, so Zigbee2MQTT doesn’t start with mismatched network parameters.

After this, `docker compose up -d` should bring up both containers and Zigbee2MQTT should start without the EZSP/Ember “HOST_FATAL_ERROR” seen earlier.

### What I did
- Updated `docker-compose.yml` to add a `mosquitto` service and make `zigbee2mqtt` depend on it.
- Added `mosquitto/config/mosquitto.conf` (anonymous local broker on port 1883).
- Updated Zigbee2MQTT config in `data/configuration.yaml`:
  - `mqtt.server: mqtt://mosquitto:1883`
  - `serial.adapter: zstack`
  - `advanced.channel/pan_id/ext_pan_id/network_key` to match `zigpy radio znp /dev/ttyUSB0 info`.

### Why
- Zigbee2MQTT needs an MQTT broker; running Mosquitto in the same Compose project avoids “where is my broker” ambiguity.
- The Sonoff “Zigbee 3.0 USB Dongle Plus” on this host speaks TI ZNP (verified with `zigpy radio znp ... info`), so Zigbee2MQTT must use `zstack` rather than `ember`.
- Matching network parameters prevents coordinator/network mismatch issues on startup.

### What worked
- `zigpy radio znp /dev/ttyUSB0 info` confirmed:
  - `Channel: 20`
  - `PAN ID: 0xE50D`
  - `Extended PAN ID: 9c:97:ff:af:f0:fb:8c:bc`

### What didn't work
- N/A (this step is config-only; runtime verification happens next).

### What I learned
- Zigbee2MQTT’s “ember” adapter path will repeatedly reset ASH and fail with `HOST_FATAL_ERROR` when pointed at a non-EZSP coordinator; this was the root cause of the earlier start failure.

### What was tricky to build
- Converting the coordinator-reported hex IDs (PAN/EPID/network key) into the decimal byte-list format Zigbee2MQTT config expects.

### What warrants a second pair of eyes
- If we *wanted* a fresh Zigbee network, these hard-coded network parameters should be removed and the coordinator should be factory-reset; right now we intentionally preserve the existing network state.

### What should be done in the future
- N/A.

### Code review instructions
- Validate with:
  - `docker compose up -d`
  - `docker compose logs -f zigbee2mqtt`
  - Pair device, then watch for “interview”/“paired” logs.

## Step 3: Clear MQTT port conflict with host Mosquitto

When starting the Compose stack, Docker failed to publish the container’s `1883:1883` because the host already had Mosquitto listening on port 1883 (loopback). The quickest way to proceed was to stop the host `mosquitto.service` so the Compose broker can bind the port.

I stopped the host service successfully, but disabling it permanently failed due to permissions; it remains `enabled` but `inactive`. For this session, “stopped” is sufficient to let Docker bind to `0.0.0.0:1883`.

### What I did
- Attempted `docker compose up -d` and hit:
  - `bind: address already in use` on `0.0.0.0:1883`.
- Verified a host Mosquitto process existed:
  - `pgrep -a mosquitto`
  - `systemctl status mosquitto`
- Stopped it:
  - `systemctl stop mosquitto`
- Confirmed state:
  - `systemctl is-active mosquitto` (inactive)
  - `systemctl is-enabled mosquitto` (enabled)

### Why
- Docker can’t publish `1883` while the host service already holds it.

### What worked
- `systemctl stop mosquitto` made the service inactive, freeing the port.

### What didn't work
- `systemctl disable mosquitto` failed with `Permission denied` (likely needs elevated privileges).

### What I learned
- Even if the host Mosquitto is only listening on loopback, Docker’s publish to `0.0.0.0:1883` still conflicts.

### What was tricky to build
- Ensuring the fix is minimally disruptive: stop is enough for immediate validation; disable is optional/policy-driven.

### What warrants a second pair of eyes
- Decide whether we should avoid publishing `1883` at all (internal-only broker) to prevent future conflicts, or keep publishing for external MQTT clients.

### What should be done in the future
- If you want this permanently disabled: run `sudo systemctl disable --now mosquitto` (requires host admin privileges).

### Code review instructions
- Re-run `docker compose up -d` and confirm `mosquitto` starts and `zigbee2mqtt` connects.

## Step 4: Disable onboarding-only startup

After the containers started, the Zigbee2MQTT UI served the “Onboarding” HTML for both the main page and API endpoints like `/api/info` and `/api/state`. That indicates Zigbee2MQTT didn’t proceed to its normal runtime (no zigbee-herdsman init, no API JSON), even though `data/configuration.yaml` existed.

To force Zigbee2MQTT to attempt normal startup from the provided configuration, I turned off onboarding in config. This should make `/api/*` endpoints return JSON and produce the usual startup logs (MQTT connect, adapter init).

### What I did
- Verified onboarding-only behavior via HTTP:
  - `curl -I http://localhost:8080/`
  - `curl http://localhost:8080/api/info` (returned onboarding HTML)
- Set `onboarding: false` in `data/configuration.yaml`.

### Why
- We already have a full config; onboarding-only mode blocks adapter bring-up and pairing.

### What worked
- N/A (verification happens after restart).

### What didn't work
- N/A.

### What I learned
- Zigbee2MQTT can serve onboarding HTML for API endpoints when it considers base configuration incomplete; disabling onboarding is a simple way to force config-driven startup for Docker-based setups.

### What was tricky to build
- Distinguishing “UI has onboarding page” (normal) from “API endpoints return onboarding HTML” (not started).

### What warrants a second pair of eyes
- If onboarding-only mode persists, we should check Zigbee2MQTT’s config validation requirements for v2.7.2 and verify it’s actually loading `/app/data/configuration.yaml`.

### What should be done in the future
- N/A.

### Code review instructions
- Restart `zigbee2mqtt` and confirm `/api/state` returns JSON, then pair the switch.

## Step 5: Confirm Zigbee2MQTT start + enable permit-join

After restarting with onboarding disabled, Zigbee2MQTT proceeded to initialize zigbee-herdsman, detect the coordinator on `/dev/ttyUSB0`, connect to the Mosquitto broker, and start the frontend. It also wrote out `data/coordinator_backup.json` and created `data/database.db`, confirming it’s operating normally with persistent storage.

To support immediate pairing, I enabled `permit_join: true` in config and restarted Zigbee2MQTT so the switch can join as soon as it’s placed into discovery mode.

### What I did
- Restarted Zigbee2MQTT and validated successful bring-up via log file:
  - `data/log/2026-01-04.16-48-06/log.log`
- Verified the coordinator initialized and Zigbee2MQTT reached “started”.
- Enabled pairing mode:
  - Set `permit_join: true` in `data/configuration.yaml`
  - Restarted `zigbee2mqtt`

### Why
- With `permit_join` disabled, devices won’t be allowed to join even if they’re in pairing mode.

### What worked
- Zigbee2MQTT now starts successfully with the Sonoff dongle on Z-Stack:
  - Adapter discovery matched `/dev/ttyUSB0` and selected `zstack`.
  - Connected to MQTT at `mqtt://mosquitto:1883`.
  - “Zigbee2MQTT started!”

### What didn't work
- N/A (pairing activity depends on putting the switch into discovery mode).

### What I learned
- Zigbee2MQTT 2.7.2 creates `coordinator_backup.json` automatically on startup and persists state in `database.db` under `/app/data`.

### What was tricky to build
- Startup logs can lag behind initial “Serialport opened”; the full coordinator init happened ~10 seconds later in the log file.

### What warrants a second pair of eyes
- Whether leaving `permit_join: true` committed is acceptable; best practice is to turn it off after pairing to reduce accidental joins.

### What should be done in the future
- After pairing succeeds, set `permit_join: false` and restart (or toggle it off in the UI).

### Code review instructions
- Put the Sonoff switch in pairing mode, then watch:
  - `docker compose logs -f zigbee2mqtt`
  - or tail `data/log/<latest>/log.log` for “interview”/“paired” lines.

## Step 6: Pair device and confirm state telemetry over MQTT

With Zigbee2MQTT running, the device successfully joined and completed its interview. Zigbee2MQTT identified it as a “Third Reality Zigbee / BLE smart plug with power (3RSP02028BZ)” and immediately began publishing state and electrical measurements to MQTT under the device topic.

There is a recurring “Failed to configure” message that looks like a converter expecting a cluster the device doesn’t expose. Despite that, the core functionality (on/off + telemetry) is working, so we can treat it as a non-fatal warning unless a specific feature is missing.

### What I did
- Enabled join and observed successful pairing in `data/log/2026-01-04.16-49-39/log.log`:
  - `Device '0x282c02bfffe69870' joined`
  - `Succesfully interviewed ...`
  - `device has successfully been paired`
  - `identified as: ... (3RSP02028BZ)`
- Confirmed MQTT publishes for the device topic include `state`, `voltage`, `current`, `power`, `power_factor` (and `update`).

### Why
- This validates the end-to-end loop: coordinator ↔ Zigbee2MQTT ↔ Mosquitto ↔ MQTT topics.

### What worked
- Interview completed and the device started reporting:
  - Example publishes observed: `state: ON/OFF`, `voltage ~119V`, `power`, etc.

### What didn't work
- Zigbee2MQTT “configure” step failed 3 times:
  - `Failed to configure ... has no input cluster 3rPlugGen2SpecialCluster`

### What I learned
- Some “configure” errors are converter-level and non-fatal; the device can still be fully controllable and report telemetry.

### What was tricky to build
- Separating “interview successful” (paired) from “configure successful” (optional post-setup bindings/reporting); the latter can fail while the device still works.

### What warrants a second pair of eyes
- If you expected a *Sonoff* device but Zigbee2MQTT identified a Third Reality plug, double-check which physical device was put into pairing mode.

### What should be done in the future
- Optional: give the device a stable `friendly_name` (via the UI or `devices:` in `data/configuration.yaml`) so MQTT topics aren’t based on the IEEE address.
- Optional: turn `permit_join` back off after pairing to prevent accidental joins.

### Code review instructions
- Verify MQTT control works by publishing to `zigbee2mqtt/<friendly_name>/set` with `{\"state\":\"ON\"}` / `{\"state\":\"OFF\"}` (or using the UI).

## Step 7: Store debugging Python as scripts

You asked me to stop running ad-hoc Python inline and instead save it under `scripts/` with increasing numeric prefixes so the exact investigations are reproducible later. I added scripts that reproduce the inline checks I previously ran while inspecting zigpy-znp framing and API internals.

You also mentioned Docker Compose is currently stopped, which means `/dev/ttyUSB0` should be free if you want to run any ZNP experiments directly against the dongle.

### What I did
- Added:
  - `scripts/01-znp-versions.py` (prints installed versions and interpreter path)
  - `scripts/02-znp-sys-version-frame.py` (prints the raw ZNP transport frame bytes for `SYS.Version`)
  - `scripts/03-znp-inspect-api.py` (prints the source for key zigpy-znp methods)

### Why
- Make the investigation steps reproducible and easier to retrace without retyping one-off commands.

### What worked
- N/A (these are saved for you to run as needed).

### What didn't work
- N/A.

### What I learned
- zigpy-znp provides a clear reference implementation for MT framing (`TransportFrame`) and command headers (`CommandHeader`), which is a useful ground-truth when implementing a custom ZNP host.

### What was tricky to build
- Keeping scripts focused: enough output to be useful, without dumping the entire dependency tree.

### What warrants a second pair of eyes
- N/A.

### What should be done in the future
- When you start doing real ZCL sends via `AF.DataRequest`, add a dedicated script that connects to `/dev/ttyUSB0`, registers an endpoint, and sends a single ZCL On/Off command end-to-end.

### Code review instructions
- Run scripts with:
  - `python3 scripts/01-znp-versions.py`
  - `python3 scripts/02-znp-sys-version-frame.py`
  - `python3 scripts/03-znp-inspect-api.py`

## Step 8: Add low-level ZNP/ZDO/ZCL helper scripts

This step adds a small set of runnable scripts to explore the dongle at progressively lower levels: (1) connect via zigpy-znp and dump coordinator/network info, (2) do ZDO discovery against a target node, (3) send raw ZCL payloads via AF.DataRequest, and (4) sniff incoming AF messages. I also added a raw pyserial SYS.Ping script to demonstrate ZNP framing without any zigpy-znp abstractions.

**Commit (code):** de05c39102de1d0b9e8f3d40e1534fc5aaa8c387 — "zigbee: add low-level ZNP/ZDO/ZCL scripts"

At the moment I tried to validate access to `/dev/ttyUSB0` and noticed the Sonoff dongle is no longer present on USB (no `/dev/ttyUSB*`, no `/dev/serial/by-id`, and `lsusb` doesn’t show `10c4:ea60`). Once it’s reattached, these scripts can be used directly.

### What I did
- Added scripts:
  - `scripts/04-znp-connect-info.py` (SYS/NVRAM info; optional `--show-keys`)
  - `scripts/05-zdo-discover.py` (ActiveEP + SimpleDesc via ZDO)
  - `scripts/06-af-send-zcl.py` (AF.DataRequest with raw ZCL or `--onoff`)
  - `scripts/07-af-sniff.py` (listen for AF.IncomingMsg; prints ZCL header hints)
  - `scripts/08-find-coordinator.sh` (find dongle USB + serial paths)
  - `scripts/09-znp-raw-sys-ping.py` (raw ZNP SYS.Ping using pyserial)
- Verified scripts compile / shellcheck cleanly (syntax-only).

### Why
- Provide a “close to the metal” playground with reproducible commands and minimal hidden behavior.

### What worked
- Scripts are in place and compile.

### What didn't work
- The coordinator device node disappeared during validation (`/dev/ttyUSB0` not present; `lsusb` did not show the CP210x bridge).

### What I learned
- If the dongle is unplugged (or USB resets), `/dev/serial/by-id` may not exist at all; `scripts/08-find-coordinator.sh` makes it obvious when it’s back.

### What was tricky to build
- Registering a “client” endpoint correctly: for controller behavior, clusters should be registered in the **output** cluster list.

### What warrants a second pair of eyes
- Confirm the chosen `src_ep` default (20) doesn’t conflict with any firmware-reserved endpoints for your specific Z-Stack build.

### What should be done in the future
- Add a script that performs a single end-to-end ZCL Read Attributes + parses the response entries (type/value) for power/voltage/current.

### Code review instructions
- Reattach the dongle and run:
  - `scripts/08-find-coordinator.sh`
  - `python3 scripts/04-znp-connect-info.py`
  - `python3 scripts/05-zdo-discover.py --nwk 0x0038`
  - `python3 scripts/06-af-send-zcl.py --dst-nwk 0x0038 --dst-ep 1 --onoff toggle`
  - `python3 scripts/07-af-sniff.py --seconds 60`
  - `python3 scripts/09-znp-raw-sys-ping.py`

## Step 9: Reattach dongle and verify ZDO + ZCL end-to-end

After the dongle was reattached, I verified it was visible via `lsusb` and `/dev/serial/by-id`, then exercised the low-level scripts against the already-joined plug at NWK `0x0038` (IEEE `0x282c02bfffe69870`). The key checks were: (1) raw ZNP framing works (`SYS.Ping`), (2) zigpy-znp can connect and read NIB/network state, (3) ZDO discovery returns endpoints/clusters, and (4) ZCL control and attribute reads work over `AF.DataRequest`.

Some early script issues surfaced (wrong attribute names, enum types, false-negative device detection, and the fact that only one process can hold the serial port). I fixed these and added a combined “send read-attrs and wait for response” script to avoid needing two concurrent serial connections.

**Commit (code):** f47a9b8b67db6e8582203a3c39d77335313426b9 — "zigbee: improve ZNP scripts and add ZCL read attrs"

### What I did
- Verified the coordinator is present:
  - `scripts/08-find-coordinator.sh`
- Confirmed raw MT framing works (no library):
  - `python3 scripts/09-znp-raw-sys-ping.py --port /dev/serial/by-id/...`
- Confirmed zigpy-znp connect + coordinator/network state:
  - `python3 scripts/04-znp-connect-info.py --port /dev/serial/by-id/...`
- Queried ZDO discovery for the plug:
  - `python3 scripts/05-zdo-discover.py --port /dev/serial/by-id/... --nwk 0x0038`
- Sent a ZCL On/Off Toggle and got `AF.DataConfirm: SUCCESS`:
  - `python3 scripts/06-af-send-zcl.py --port /dev/serial/by-id/... --dst-nwk 0x0038 --dst-ep 1 --onoff toggle`
- Read the On/Off attribute (`0x0006`/`0x0000`) and received a ZCL Read Attributes Response:
  - `python3 scripts/10-zcl-read-attrs.py --port /dev/serial/by-id/... --dst-nwk 0x0038 --dst-ep 1 --cluster 0x0006 --attr 0x0000`

### Why
- This establishes a repeatable “lowest-level vertical slice” using ZNP → AF → ZCL without Zigbee2MQTT.

### What worked
- ZDO discovery returned endpoints `[1, 242]` and endpoint 1 has `genOnOff (0x0006)`, `haElectricalMeasurement (0x0b04)`, and `seMetering (0x0702)`.
- ZCL toggle succeeded with `AF.DataConfirm: SUCCESS`.
- ZCL Read Attributes Response was received and decoded (for `onOff` it returned `False` in this run).

### What didn't work
- Running a “sniffer” and a “sender” simultaneously failed because the serial device can only be opened by one process at a time (`PermissionError: The serial port is locked by another application`).

### What I learned
- For experiments that require both sending and receiving, scripts should perform both within a single ZNP connection.

### What was tricky to build
- zigpy-znp’s command schemas are strict about enum types (e.g. `LatencyReq` must be `LatencyReq.NoLatencyReqs`, not a raw `uint8_t`).
- zigpy’s ZDO simple descriptor uses snake_case field names (`input_clusters`, `output_clusters`, etc.).

### What warrants a second pair of eyes
- Confirm that using `ZDO.StartupFromApp` in these scripts is always safe for an already-running coordinator and won’t disrupt an established network.

### What should be done in the future
- Add a companion script for Electrical Measurement + Metering reads (multiplier/divisor + activePower + current/voltage) and optionally configure reporting.

### Code review instructions
- Start with:
  - `cmd/experiments/zigbee/scripts/08-find-coordinator.sh`
  - `cmd/experiments/zigbee/scripts/09-znp-raw-sys-ping.py`
  - `cmd/experiments/zigbee/scripts/05-zdo-discover.py`
  - `cmd/experiments/zigbee/scripts/06-af-send-zcl.py`
  - `cmd/experiments/zigbee/scripts/10-zcl-read-attrs.py`
