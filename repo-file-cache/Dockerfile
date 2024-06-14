FROM openeuler/openeuler:23.03 as BUILDER
RUN dnf update -y && \
    dnf install -y golang && \
    go env -w GOPROXY=https://goproxy.cn,direct

MAINTAINER zengchen1024<chenzeng765@gmail.com>

# build binary
COPY . /go/src/github.com/opensourceways/repo-file-cache
RUN cd /go/src/github.com/opensourceways/repo-file-cache && CGO_ENABLED=1 go build -v -o ./repo-file-cache main.go

# copy binary config and utils
FROM openeuler/openeuler:22.03
RUN dnf -y update && \
    dnf in -y shadow && \
    groupadd -g 1000 repo-file-cache && \
    useradd -u 1000 -g repo-file-cache -s /bin/bash -m repo-file-cache && \
    mkdir -p /opt/app/

USER repo-file-cache

COPY --chown=repo-file-cache ./conf /opt/app/conf
# overwrite config yaml
COPY --chown=repo-file-cache --from=BUILDER /go/src/github.com/opensourceways/repo-file-cache/repo-file-cache /opt/app

WORKDIR /opt/app/
ENTRYPOINT ["/opt/app/repo-file-cache"]
