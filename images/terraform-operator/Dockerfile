FROM alpine as build

ADD https://releases.hashicorp.com/terraform/0.13.5/terraform_0.13.5_linux_amd64.zip /files/terraform.zip

WORKDIR /files
RUN mkdir /tool
RUN for f in *.zip; do unzip $f -d /tool; done

FROM alpine

# RUN apk add --no-cache -U git openssh
COPY --from=build /tool/* /usr/local/bin/

ENV HOME=/opt/terraform \
    USER_NAME=terraform \
    USER_UID=1001

RUN echo "${USER_NAME}:x:${USER_UID}:0:${USER_NAME} user:${HOME}:/sbin/nologin" >> /etc/passwd

WORKDIR ${HOME}
USER ${USER_UID}

ARG BIN=bin/terraform-operator
COPY $BIN /usr/local/bin/terraform-operator

ENTRYPOINT ["/usr/local/bin/terraform-operator", "run", "--watches-file=./watches.yaml"]
