FROM mongo:4.0.4

ENV AUTH yes
ENV STORAGE_ENGINE wiredTiger
ENV JOURNALING no

ADD run.sh /run.sh
ADD set_mongodb_password.sh /set_mongodb_password.sh

CMD ["/run.sh"]