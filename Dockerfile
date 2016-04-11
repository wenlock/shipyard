FROM golang:1.5

ENV PATH=$PATH:/go/bin:/usr/local/go/bin \
    GO15VENDOREXPERIMENT=1 \
    SY_DIR=/go/src/github.com/shipyard/shipyard \
    SY_USER=sy_user

COPY . $SY_DIR
WORKDIR $SY_DIR

RUN apt-get update && \
    curl -sL https://deb.nodesource.com/setup_5.x | bash - && \
    apt-get install -y nodejs && \
    apt-get clean && \
    npm install -g bower && \
    go get github.com/aktau/github-release

RUN useradd -b /home -u 1000 -m -s /bin/bash $SY_USER && \
    chown -R $SY_USER:$SY_USER /home/$SY_USER && \
    chown -R $SY_USER:$SY_USER $SY_DIR && \
    make all

USER $SY_USER

EXPOSE 8080

CMD ["./shipyard"]
