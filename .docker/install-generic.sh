#!/bin/sh
export $(cat /tmp/build.env | xargs)
set -ex

