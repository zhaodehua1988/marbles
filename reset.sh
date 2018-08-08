#!/bin/bash
cd ./network 
 ./stop.sh 
 ./teardown.sh 
 ./startFabric.sh \
&& cd .. \
&& node ./scripts/install_chaincode.js \
&& node ./scripts/instantiate_chaincode.js \
&& gulp marbles_local