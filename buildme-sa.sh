#!/bin/bash -x
CURRENT_TIMESTAMP=$(date +%s)
SY_CONTAINER_NAME="shipyard-$CURRENT_TIMESTAMP"
# RDB_CONTAINER_NAME="rethink-$CURRENT_TIMESTAMP"
#SY_IMAGE_NAME="$SY_CONTAINER_NAME"
SY_IMAGE_NAME=controller_sa

MY_BUILD_PROXY="--build-arg http_proxy=$http_proxy --build-arg https_proxy=$http_proxy --build-arg no_proxy=$no_proxy"
#MY_ENV_PROXY=$(echo $MY_BUILD_PROXY | sed "s/build-arg/env/g")

docker build $MY_BUILD_PROXY --tag $SY_IMAGE_NAME -f Dockerfile-sa.build .

#docker run -i --name $SY_CONTAINER_NAME $MY_ENV_PROXY -v /DATA/ilm_repo:/go/src/github.com/shipyard/shipyard --entrypoint=/bin/bash $SY_IMAGE_NAME -c "make clean && make all"
result=$?


#docker rm -fv $SY_CONTAINER_NAME && 
#docker rmi -f $SY_IMAGE_NAME

exit $result

