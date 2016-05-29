#!/bin/bash
#***************************************************************************************
# This script builds tests and deploys ILM
# View commands results by setting up an environment variable named ILM_DEBUG with any value
#***************************************************************************************

# Any of these parameters can be predefined in the environment before runnign the script
SY_IMAGE_NAME=${SY_IMAGE_NAME:-"ilmhpe/ilm:dev"}
DOCKERFILE=${DOCKERFILE:-"Dockerfile-sa.build"}
ILM_DEV_IMAGE=${ILM_DEV_IMAGE:-"ilmhpe/ilm-go"}
ILM_NO_PROXY=${ILM_NO_PROXY:-"swarm,proxy,controller,rethinkdb,postgres,clair"}
COMPOSE_FILE=${COMPOSE_FILE:-"docker-compose.yml"}
RETHINK_IMAGE=${RETHINK_IMAGE:-rethinkdb:2.3}
ILM_DEBUG=${ILM_DEBUG:-""}
# For ILM tests only
# Supply rethinkdb data folder, if needed
RETHINKDB_DATA_FOLDER=${RETHINKDB_DATA_FOLDER:-""}
RDB_CONTAINER_NAME=${RDB_CONTAINER_NAME:-"rethinkdb_$(date +%Y%m%d_%H%M%S)"}

# Proxy settings are read off the system. Override with care!
PROXY_DETECTED=${PROXY_DETECTED:-""}
WITH_PROXY=${WITH_PROXY:-""}
USAGE="$0 [OP] {-d}\nWhere OP is one of: all build test deploy clean down revert\nSpecify -d to start deployment in daemon mode."


# TODO: Add "--user $(id -u)" 
# This will keep all files generated owned by the same user, allowing for easier cleanup and consistency
# But first fix issue "Error: EACCES, permission denied '/.config'" (likely with packaging controller into shipyard executable)
# NOte: whoami does not work in docker 1.11.1: https://github.com/docker/docker/issues/22323
ILM_BUILD_CMD="docker run --rm -i -v $(pwd):/go/src/github.com/shipyard/shipyard \
--entrypoint=/bin/bash"

if [ "$2" == "-d" ]; then
    AS_DAEMON=$2
elif [ -z "$2" ]; then
    AS_DAEMON=
else 
    echo "Incorrect usage: Malformed second parameter"
    echo "$USAGE"
    echo "staring in normal (non-daemon) mode"
fi

#******************************************************************************
# prepare the .env file and prepare the docker build proxy settings
function SET_ILM_PROXY_SETTING() {
    if [ ! -z "$WITH_PROXY" ]; then
        if [ ! -z "$http_proxy" ]; then
          echo "http_proxy=$http_proxy" > .env
          PROXY_DETECTED="yes"
          WITH_PROXY="--build-arg \"http_proxy=$http_proxy\""
        fi

        if [ ! -z "$https_proxy" ]; then
          echo "https_proxy=$https_proxy" >> .env
          WITH_PROXY="$WITH_PROXY --build-arg \"https_proxy=$https_proxy\""
        elif [ ! -z $PROXY_DETECTED ]; then
          echo "https_proxy=$http_proxy" >> .env
          WITH_PROXY="$WITH_PROXY --build-arg \"https_proxy=$http_proxy\""
        fi

        if [ ! -z "$no_proxy" ]; then
          if [ ! -z "$PROXY_DETECTED" ]; then
            echo "no_proxy=$ILM_NO_PROXY,$no_proxy" >> .env
            WITH_PROXY="$WITH_PROXY --build-arg \"no_proxy=$ILM_NO_PROXY,$no_proxy\""
          else
            echo "no_proxy=$ILM_NO_PROXY" >> .env
            WITH_PROXY="$WITH_PROXY --build-arg \"no_proxy=$ILM_NO_PROXY\""
          fi
        fi
    fi
}

#******************************************************************************
# build the binaries
function ILM_BUILD() {
    echo "   Building ILM..."
    echo "$ILM_BUILD_CMD $ILM_DEV_IMAGE -c 'make clean && make all'"
    $ILM_BUILD_CMD $ILM_DEV_IMAGE -c 'make clean && make all'
    result=$?

    if [ $result -ne 0 ]; then
        echo "   Error: Could not build ILM! Exiting."
        exit $result
    fi
}

#******************************************************************************
# run tests
function ILM_TEST() {
    echo "   Testing ILM..."
    echo "   Starting a rethinkdb instance"
    MY_CMD="docker run -d --name $RDB_CONTAINER_NAME"

    if [ ! -z "$RETHINKDB_DATA_FOLDER" ]; then
        MY_CMD="$MY_CMD -v $RETHINKDB_DATA_FOLDER:/data"
    fi
    MY_CMD="$MY_CMD $RETHINK_IMAGE"
    echo "$MY_CMD"
    $MY_CMD
    result=$?
    if [ $result -ne 0 ]; then
        echo "   Error: Could not start required rethinkdb container."
        exit $result
    fi

    echo "   Starting ILM tests..."
    echo "$ILM_BUILD_CMD --link $RDB_CONTAINER_NAME:rethinkdb $ILM_DEV_IMAGE -c 'make test'"
    $ILM_BUILD_CMD --link $RDB_CONTAINER_NAME:rethinkdb $ILM_DEV_IMAGE -c 'make test'
    result=$?
    
    # remove the rethinkdb container
    docker rm -f $RDB_CONTAINER_NAME

    if [ $result -ne 0 ]; then
        echo "   Error: Could not test ILM! Exiting."
        exit $result
    fi
}

#******************************************************************************
# clean the binaries
function ILM_CLEAN() {
    echo "   Cleaning ILM..."
    echo "$ILM_BUILD_CMD $ILM_DEV_IMAGE -c 'make clean'"
    $ILM_BUILD_CMD $ILM_DEV_IMAGE -c 'make clean'
    result=$?
    
    if [ $result -ne 0 ]; then
        echo "   Error: Could not clean ILM! Exiting."
        exit $result
    fi
}

#******************************************************************************
# build a new image with the compiled binaries
function CREATE_IMAGE() {
    echo "   Creating new ILM image: $SY_IMAGE_NAME"
    echo "Removing existing image first..."
    docker rmi $SY_IMAGE_NAME
    # ignoring error for now (likely image does not exist locally yet)

    echo "docker build $WITH_PROXY --tag $SY_IMAGE_NAME -f $DOCKERFILE $(pwd)"
    docker build $WITH_PROXY --tag $SY_IMAGE_NAME -f $DOCKERFILE $(pwd)
    result=$?

    if [ $result -ne 0 ]; then
        echo "   Error: Could not build a new ILM image! Exiting."
        exit $result
    fi
}


#******************************************************************************
function UPDATE_COMPOSE_FILE() {
    echo "   Updating compose file..."
    sed -i -e "s;image: ilmhpe/ilm.*;image: $SY_IMAGE_NAME;" $COMPOSE_FILE 
    result=$?

    if [ $result -ne 0 ]; then
        echo "   Error: Could not update the compose file: $COMPOSE_FILE. Exiting."
        exit $result
    fi
}

#******************************************************************************
function REVERT_COMPOSE_FILE() {
    echo "   Reverting changes to compose file..."
    sed -i -e "s;image: $SY_IMAGE_NAME;image: ilmhpe/ilm;" $COMPOSE_FILE
    result=$?

    if [ $result -ne 0 ]; then
        echo "   Error: Could not restore the compose file: $COMPOSE_FILE. Exiting."
        exit $result
    fi
}

#******************************************************************************
function DEPLOY_COMPOSE() {
    echo "   Deploying ILM..."
    echo "docker-compose -f $COMPOSE_FILE up $AS_DAEMON"
    docker-compose -f $COMPOSE_FILE up $AS_DAEMON
    result=$?

    if [ $result -ne 0 ]; then
        echo "    Error: Could not deploy ILM."
        exit $result
    fi
}

#******************************************************************************
function DEPLOY_TEARDOWN() {
    echo "   Taking down existing ILM deployment..."
    echo "docker-compose -f $COMPOSE_FILE down"
    docker-compose -f $COMPOSE_FILE down
    result=$?

    if [ $result -ne 0 ]; then
        echo "    Error: Could not teardown ILM."
        exit $result
    fi
}
#******************************************************************************
case "$1" in
    test)
        ILM_TEST
        ;;
    clean)
        ILM_CLEAN
        REVERT_COMPOSE_FILE
        ;;
    build)
        SET_ILM_PROXY_SETTING
        ILM_BUILD
        CREATE_IMAGE
        ;;
    deploy)
        DEPLOY_TEARDOWN
        SET_ILM_PROXY_SETTING
        UPDATE_COMPOSE_FILE
        DEPLOY_COMPOSE
        ;;
    all)
        DEPLOY_TEARDOWN
        SET_ILM_PROXY_SETTING
        ILM_BUILD
        CREATE_IMAGE
        UPDATE_COMPOSE_FILE
        DEPLOY_COMPOSE
        ;;
    down)
        DEPLOY_TEARDOWN
        ;;
    revert)
        REVERT_COMPOSE_FILE
        ;;
    *)
        echo -e $USAGE
        exit 1
esac

echo "Done!"
exit 0