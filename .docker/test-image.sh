#!/bin/sh

imageName=${1}
if test -z ${imageName}; then
    echo "Usage: $0 <imageName>" 1>&2
    exit 1
fi
export $(cat .docker/build.env | xargs)

set -ex

docker run --rm ${imageName} goxr --version 2>&1          | grep "Version:      TEST${TRAVIS_BRANCH}TEST"
docker run --rm ${imageName} goxr --version 2>&1          | grep "Git revision: TEST${TRAVIS_COMMIT}TEST"
docker run --rm ${imageName} goxr-server --version 2>&1   | grep "Version:      TEST${TRAVIS_BRANCH}TEST"
docker run --rm ${imageName} goxr-server --version 2>&1   | grep "Git revision: TEST${TRAVIS_COMMIT}TEST"
