#!/bin/bash

PROXY_DETECTED=
ILM_NO_PROXY="swarm,proxy,controller,rethinkdb,postgres,clair"

if [ ! -z "$http_proxy" ]; then
  echo "http_proxy=$http_proxy" > .env
  PROXY_DETECTED="yes"
fi

if [ ! -z "$https_proxy" ]; then
  echo "https_proxy=$https_proxy" >> .env
elif [ ! -z $PROXY_DETECTED ]; then
  echo "https_proxy=$http_proxy" >> .env
fi

if [ ! -z "$no_proxy" ]; then
  if [ ! -z "$PROXY_DETECTED" ]; then
    echo "no_proxy=$ILM_NO_PROXY,$no_proxy" >> .env
  else
    echo "no_proxy=$ILM_NO_PROXY" >> .env
  fi
fi

exit 0