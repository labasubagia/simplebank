#!/bin/bash

# Run this from workspasceRoot
# See ../../Makefile

# build
go build

# create workdir
mkdir -p /opt/simplebank/db
cp -R db/migration /opt/simplebank/db/
cp -u simplebank /opt/simplebank/simplebank
cp -u infra/systemd/.env.systemd /opt/simplebank/.env

# add systemd config
cp -u infra/systemd/simplebank.service /lib/systemd/system/

# run service
systemctl start simplebank
# systemctl enable simplebank # uncomment for startup
systemctl status simplebank
