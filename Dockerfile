FROM ubuntu:latest

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -yq iproute2

RUN useradd -rm -d /home/coap -s /bin/bash -u 1002 coap

COPY --chown=coap:coap ./server/coap-server         /home/coap/coap-server
COPY --chown=coap:coap ./server/start.sh            /home/coap/start.sh
COPY --chown=coap:coap ./certificates/CA.crt        /home/coap/CA.crt
COPY --chown=coap:coap ./certificates/server.crt    /home/coap/server.crt
COPY --chown=coap:coap ./certificates/server.key    /home/coap/server.key

USER coap
WORKDIR /home/coap

EXPOSE 5688/udp

CMD [ "/home/coap/start.sh" ]