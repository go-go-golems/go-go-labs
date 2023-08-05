# GO GO LABS

```
 _______  _______    _______  _______    ___      _______  _______  _______ 
|       ||       |  |       ||       |  |   |    |   _   ||  _    ||       |
|    ___||   _   |  |    ___||   _   |  |   |    |  |_|  || |_|   ||  _____|
|   | __ |  | |  |  |   | __ |  | |  |  |   |    |       ||       || |_____ 
|   ||  ||  |_|  |  |   ||  ||  |_|  |  |   |___ |       ||  _   | |_____  |
|   |_| ||       |  |   |_| ||       |  |       ||   _   || |_|   | _____| |
|_______||_______|  |_______||_______|  |_______||__| |__||_______||_______|
```

Assorted software ideas that might mean something or not. Some of them will turn into proper tools,
the rest is just temporary garbo.

Useful:

- cmd/excel2json - does what it says
- cmd/gtm - extract and represent data about variables, triggers and tags from a google tag manager container
- cmd/mastoid - download and render conversation threads from mastodon for archival

Experiments:

- cmd/aipl - try to parse the AIPL language as a frontend for geppetto
- cmd/monads - experiment with different monads in golang
- cmd/weave - interface to the weaviate database

WIP:

- reggie - run regexps against text files

## Installation

To install the `mastoid` command line tool with homebrew, run:

```bash
brew tap go-go-golems/go-go-go
brew install go-go-golems/go-go-go/go-go-labs
```

To install the `mastoid` command using apt-get, run:

```bash
echo "deb [trusted=yes] https://apt.fury.io/go-go-golems/ /" >> /etc/apt/sources.list.d/fury.list
apt-get update
apt-get install go-go-labs
```

To install using `yum`, run:

```bash
echo "
[fury]
name=Gemfury Private Repo
baseurl=https://yum.fury.io/go-go-golems/
enabled=1
gpgcheck=0
" >> /etc/yum.repos.d/fury.repo
yum install go-go-labs
```

To install using `go get`, run:

```bash
go get -u github.com/go-go-golems/go-go-labs/cmd/mastoid
```

Finally, install by downloading the binaries straight from [github](https://github.com/go-go-golems/go-go-labs/releases).

## Usage

### Registering an app

Before starting to use mastoid, you need to register an app against your server and obtain
an access token (replace https://hachyderm.io/ with your server):

```
❯ mastoid register --server https://hachyderm.io/
App registration successful!
Client ID: FOO
Client Secret: BAR
Auth URI: ...
Redirect URI: urn:ietf:wg:oauth:2.0:oob
Grant Token: ...
Access Token: ...
Website: 
Name: mastoid
Grant Token: ...
Access Token: ...

```

This will create a `~/.mastoid/config.yaml` file storing all your credentials.

### Downloading a thread

To download a mastodon thread, run:

```
❯ mastoid thread -s https://hachyderm.io/@mnl/110838692946216618 --output markdown 
> Author: mnl (2023-08-05 15:13:35.024 +0000 UTC)
> URL: https://hachyderm.io/@mnl/110837656126482103
> Author URL: https://hachyderm.io/@mnl
> 
> guess i'm gonna do evernote to obsidian export/import in hard mode under linux. wish me luck...
> #obsidian #evernote
> 
> > Author: mnl (2023-08-05 19:37:15.635 +0000 UTC)
> > 
> > test
> > 
> > > Author: mnl (2023-08-05 19:37:41.649 +0000 UTC)
> > > 
> > > test1.3
> > > 
> > > > Author: neingeist@mastodon.social (2023-08-05 19:39:00 +0000 UTC)
> > > > 
> > > > @mnl test1.4
> > > > 
> > > > > Author: mnl (2023-08-05 19:44:12.18 +0000 UTC)
> > > > > 
> > > > > @neingeist test1.4.1
> > > > > 
> > > Author: mnl (2023-08-05 19:37:30.627 +0000 UTC)
> > > 
> > > test1.2
> > > 
> > > > Author: mnl (2023-08-05 19:41:24.651 +0000 UTC)
> > > > 
> > > > test2.1
> > > > 
> > > Author: mnl (2023-08-05 19:37:23.793 +0000 UTC)
> > > 
> > > test1.1
> > > 
```

You can use text, markdown or json for the output (HTML is not implemented yet).

Private thread download is still in WIP.

---

```
 _______  _______    _______  _______ 
|       ||       |  |       ||       |
|    ___||   _   |  |    ___||   _   |
|   | __ |  | |  |  |   | __ |  | |  |
|   ||  ||  |_|  |  |   ||  ||  |_|  |
|   |_| ||       |  |   |_| ||       |
|_______||_______|  |_______||_______|
 _______  _______  ___      _______  __   __  _______ 
|       ||       ||   |    |       ||  |_|  ||       |
|    ___||   _   ||   |    |    ___||       ||  _____|
|   | __ |  | |  ||   |    |   |___ |       || |_____ 
|   ||  ||  |_|  ||   |___ |    ___||       ||_____  |
|   |_| ||       ||       ||   |___ | ||_|| | _____| |
|_______||_______||_______||_______||_|   |_||_______|
 _______  __   __  ___   ___      ______  
|  _    ||  | |  ||   | |   |    |      | 
| |_|   ||  | |  ||   | |   |    |  _    |
|       ||  |_|  ||   | |   |    | | |   |
|  _   | |       ||   | |   |___ | |_|   |
| |_|   ||       ||   | |       ||       |
|_______||_______||___| |_______||______| 
 ___      _______  _______  _______    _______  _______ 
|   |    |   _   ||  _    ||       |  |       ||       |
|   |    |  |_|  || |_|   ||  _____|  |_     _||   _   |
|   |    |       ||       || |_____     |   |  |  | |  |
|   |___ |       ||  _   | |_____  |    |   |  |  |_|  |
|       ||   _   || |_|   | _____| |    |   |  |       |
|_______||__| |__||_______||_______|    |___|  |_______|
 __   __  __    _  ___      _______  _______  ___   _ 
|  | |  ||  |  | ||   |    |       ||       ||   | | |
|  | |  ||   |_| ||   |    |   _   ||       ||   |_| |
|  |_|  ||       ||   |    |  | |  ||       ||      _|
|       ||  _    ||   |___ |  |_|  ||      _||     |_ 
|       || | |   ||       ||       ||     |_ |    _  |
|_______||_|  |__||_______||_______||_______||___| |_|
 _______  __   __  _______ 
|       ||  | |  ||       |
|_     _||  |_|  ||    ___|
  |   |  |       ||   |___ 
  |   |  |       ||    ___|
  |   |  |   _   ||   |___ 
  |___|  |__| |__||_______|
 _______  _______  _______  _______  __    _  _______  ___   _______  ___     
|       ||       ||       ||       ||  |  | ||       ||   | |   _   ||   |    
|    _  ||   _   ||_     _||    ___||   |_| ||_     _||   | |  |_|  ||   |    
|   |_| ||  | |  |  |   |  |   |___ |       |  |   |  |   | |       ||   |    
|    ___||  |_|  |  |   |  |    ___||  _    |  |   |  |   | |       ||   |___ 
|   |    |       |  |   |  |   |___ | | |   |  |   |  |   | |   _   ||       |
|___|    |_______|  |___|  |_______||_|  |__|  |___|  |___| |__| |__||_______|
 _______  _______ 
|       ||       |
|   _   ||    ___|
|  | |  ||   |___ 
|  |_|  ||    ___|
|       ||   |    
|_______||___|    
 _______  _______  _______  __   __  __    _  _______  ___      _______ 
|       ||       ||       ||  | |  ||  |  | ||       ||   |    |       |
|_     _||    ___||       ||  |_|  ||   |_| ||   _   ||   |    |   _   |
  |   |  |   |___ |       ||       ||       ||  | |  ||   |    |  | |  |
  |   |  |    ___||      _||       ||  _    ||  |_|  ||   |___ |  |_|  |
  |   |  |   |___ |     |_ |   _   || | |   ||       ||       ||       |
  |___|  |_______||_______||__| |__||_|  |__||_______||_______||_______|
 _______  __   __       
|       ||  | |  |      
|    ___||  |_|  |      
|   | __ |       |      
|   ||  ||_     _| ___  
|   |_| |  |   |  |   | 
|_______|  |___|  |___| 
```
