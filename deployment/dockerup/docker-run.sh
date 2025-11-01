#!/bin/bash
docker run --rm --network moca-network -v ./deployment/dockerup/:/root/.moca-cmd mocachain/moca-cmd $1
