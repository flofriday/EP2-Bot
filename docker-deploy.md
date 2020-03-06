# Deploy with docker

## Requirements
Obviously you need to install docker for this to work.
```bash
curl -sSL https://get.docker.com | sh
```

## Deploy
```
# You only need to do this once, not when you update the bot.
docker volume create ep2bot-volume

# This step has to be done every time
docker build -t ep2bot-template .
docker run -d --restart unless-stopped
      --env TELEGRAM_TOKEN=XXXX
      --env TELEGRAM_ADMIN=YYYY
      --env GIT_URL=https://USER:PASSWORD@b3.complang.tuwien.ac.at/ep2/2020s/uebung/USER.git
      --mount type=volume,source=ep2bot-volume,target=/app/data
      --name ep2bot-container ep2bot-template

# To stop the bot again type
docker stop ep2bot-container
```
As shown in the README.md you need to replace XXXX, YYYY and USER and PASSWORD in the commands above
## Uninstall
```
docker rmi ep2bot-container ep2bot-volume
```
