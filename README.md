# EP2-Bot
A bot for "Einführung in die Programmierung 2" TU Vienna


At [TU Vienna](https://www.tuwien.at/en/) we get a git repository for "Einführung in die Programmierung 2" (Introduction to Programming 2). 
All of our assignments will get to us via that git repo and also our points.

While git is amazing I would love that data without needing to log in every time. 
So for this reason I wrote this, bot to tell me when new assignments are out.

## Try the bot
You need to install the [golang compiler](https://golang.org/).

Than type:
```bash
go build
TELEGRAM_TOKEN=XXXX \
TELEGRAM_ADMIN=YYYY \
GIT_URL=https://USER:PASSWORD@b3.complang.tuwien.ac.at/ep2/2020s/uebung/USER.git \
./EP2-Bot
```
Replace the XXXX with the token for your telegrambot (you can get this via botfather). YYYY is your Telegram user id. 
In the GIT_URL the USER is your Matrikelnumber and PASSWORD your git password.

## Deploy 
You can deploy the bot with docker (~or as a systemd service~ actually you can only deploy it via docker at the moment). 

Look at docker-deploy.md to see how you can deploy it.