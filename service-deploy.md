# Deploy as a systemd service

## Requirements
This only works on linux. 
You will need to install the go compiler.

Note: This is not as nice as the docker deployment so the docker deployment is recommended.

## Start the service
```bash
go build

# Edit ep2bot.service all the values you need to add start with a !
# You need to do this manually

sudo cp ep2bot.service /etc/systemd/system/ep2bot.service
sudo systemctl daemon-reload
sudo systemctl enable filmresourcebot
sudo systemctl start filmresourcebot
```