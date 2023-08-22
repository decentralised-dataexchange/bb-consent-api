FROM golang:1.18


WORKDIR /go/src/github.com/bb-consent/api
COPY go.mod .
COPY go.sum .

# Install go dependencies
RUN go mod download

ENV TERM xterm-256color
RUN echo 'export PS1="${debian_chroot:+($debian_chroot)}\[\033[01;32m\]\u@\h\[\033[00m\]:\[\033[01;34m\]\w\[\033[00m\]\$ "' >> ~/.bashrc
# Cleanup
RUN apt-get clean && \
    rm -rf /tmp/* /var/tmp/* && \
    rm -rf /var/lib/apt/lists/*
