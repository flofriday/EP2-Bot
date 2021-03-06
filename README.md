# EP2-Bot
A telegram bot for "Einführung in die Programmierung 2" TU Vienna

![Screenshot](screenshot.png)

~~**Try the bot on telegram: https://t.me/ep2_bot**~~ <br>
**Note:** I am no longer in this course so I disabled the bot. If you want to run a
copy of the bot, I am more than willing to help you set it up.

At [TU Vienna](https://www.tuwien.at/en/) we get a git repository for "Einführung in die Programmierung 2" (Introduction to Programming 2). 
All of our assignments get to us via that git repository.

While git is amazing I would love that information without needing to log in every time. 
So for this reason I wrote this Telegram bot to tell me when new assignments are out.
The bot crawls the repository every 30minutes and writes me and some classmates a telegram message if there are new commits that are not
from me.

## Run the bot
You need to install the [golang compiler](https://golang.org/).

Than type:
```bash
go build
TELEGRAM_TOKEN=XXXX \
TELEGRAM_ADMIN=YYYY \
GIT_URL=https://USER:PASSWORD@b3.complang.tuwien.ac.at/ep2/2020s/uebung/USER.git \
./EP2-Bot
```
Replace the XXXX with the token for your telegram bot (you can get this via [BotFather](https://t.me/BotFather)). YYYY is your telegram user id 
(you can get it via [userinfobot](https://t.me/userinfobot)). 
In the GIT_URL the USER is your student number (german: Matrikelnummer) and PASSWORD is a personal access token (you can get this in GitLab under Profile -> Settings.
Give the token only access to "read_repository") [More about access tokens](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html).

### Or run with docker
First install [docker](https://www.docker.com/)
```bash
docker build -t ep2bot-template .
docker run --rm \
      --env TELEGRAM_TOKEN=XXXX \
      --env TELEGRAM_ADMIN=YYYY \
      --env GIT_URL=https://USER:PASSWORD@b3.complang.tuwien.ac.at/ep2/2020s/uebung/USER.git \
      --name ep2bot-container \
      ep2bot-template 
```

## Deploy 
You can deploy the bot with docker or as a systemd service. (The docker deployment is recommended as it is easier.)

Look at docker-deploy.md or service-deploy.md to see how you can deploy it.