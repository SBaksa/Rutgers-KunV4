# Rutgers-KunV4

> Rutgers-kun is a Discord bot dedicated to providing utility and a sense of connectedness between Discord servers under Rutgers University. V4 is a full rewrite of [V3](https://github.com/sriRacha21/Rutgers-kun3) in Go, built for improved concurrency and performance across multiple servers.

## Why Go?

V3 was built in JavaScript (Discord.js). As the bot grew to serve more servers simultaneously, we needed better concurrency. Go's goroutine model lets the bot handle thousands of concurrent messages across servers with minimal overhead — each Discord event is processed in its own goroutine, and the command processor uses a worker pool to keep things efficient.

## Features

- **NetID verification** — 2-step email verification tied to Rutgers scarletmail, assigns roles on completion
- **Course lookup** — pull up Rutgers class info (`!course 198:112`)
- **Moderation** — slur detection with auto-delete and mod log, ignore list
- **Quotes** — save and recall quotes per user, cross-server
- **Word tracking** — track how many times you say a specific word
- **Custom commands** — per-server custom commands
- **Fun commands** — `!8ball`, `!love`, `!meow`, `!woof`, `!roll`
- **Config** — welcome channel, welcome text, log channel, agreement channel and roles

## Run Locally

1. Clone the repo
2. Copy `.env.example` to `.env` and fill in your bot token and SMTP credentials
3. Build and run:
```bash
go build -o rutgers-kun4
./rutgers-kun4
```

## Contact

- Rutgers Mathcord Mod Community or @stveb on discord
