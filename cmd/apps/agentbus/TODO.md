- no channel for the speak, one is enough, but each message should have a (optional) topic slug

- for knowledge, add a way to list keys
- jot / announce / satisfy / await should also publish to the channel automatically

- also implementation documentation detailing how this is built on top of redis

---

- each command should show which messages have been received since the last time the agent got  messages, that wait agents don't have to poll themselves if not necessary.

- store which agent jotted down something (-f not already)
- store date for jots and channel speak

- add a "monitor" verb that monitors all of these resources and prints them out in real time (using a BareCommand)
- add zerolog logging and log everything in to /tmp/agentbus.log 

- Figure out why announce might be hanging, and also add debug logging for deubgging where announce seems to hang.
- add clear / reset command 
- add something like project prefix (required as ENV variable) to avoid cross contamination


--- 

- Add a default timeout for announcing/joining
- Add glazed helpsystem like in glaze/main to register the docs (see glazed/pkg/doc as well, using go:embed)

---

, wit a set of predetermined topics slugs.
    - exchanging general information / chatter
    - task related stuff: starting / creating new tasks / in progress messages / updates / finishing
    - announcing new knowledge


---

Add a request/response wait for response verb to have blocking chatter