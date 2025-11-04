FROM openeuler/openeuler:23.03 as BUILDER
RUN dnf update -y && \
    dnf install -y wget tar gcc && \
    wget https://mirrors.aliyun.com/golang/go1.24.1.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.24.1.linux-amd64.tar.gz && \
    export PATH=$PATH:/usr/local/go/bin  && \
    echo "PATH=\$PATH:/usr/local/go/bin" >> /etc/profile && \
    go version && \
    go env -w GOPROXY=https://goproxy.cn,direct

MAINTAINER TommyLike<tommylikehu@gmail.com>

# build binary
COPY . /go/src/github.com/opensourceways/hot-topic-website-backend
RUN cd /go/src/github.com/opensourceways/hot-topic-website-backend && GO111MODULE=on /usr/local/go/bin/go build -o hot-topic-manager -buildmode=pie --ldflags "-s -linkmode 'external' -extldflags '-Wl,-z,now'"

# copy binary config and utils
FROM openeuler/openeuler:22.03
RUN groupadd -g 1000 app && \
    useradd -u 1000 -g app -s /sbin/nologin -m app

RUN echo > /etc/issue && echo > /etc/issue.net && echo > /etc/motd
RUN mkdir /home/app -p
RUN chmod 700 /home/app
RUN chown app:app /home/app

RUN echo 'set +o history' >> /root/.bashrc
RUN sed -i 's/^PASS_MAX_DAYS.*/PASS_MAX_DAYS   90/' /etc/login.defs
RUN rm -rf /tmp/*

USER app
WORKDIR /home/app

COPY --chown=app --from=BUILDER /go/src/github.com/opensourceways/hot-topic-website-backend/hot-topic-manager /home/app

RUN chmod 550 /home/app/hot-topic-manager

RUN echo "umask 027" >> /home/app/.bashrc
RUN echo 'set +o history' >> /home/app/.bashrc

ENTRYPOINT ["/home/app/hot-topic-manager"]