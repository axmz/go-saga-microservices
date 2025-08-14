FROM debezium/connect:2.7.2.Final
USER root
RUN microdnf install -y curl jq && microdnf clean all
COPY register-and-run.sh /usr/local/bin/register-and-run.sh
RUN chmod +x /usr/local/bin/register-and-run.sh
USER kafka
ENTRYPOINT ["/usr/local/bin/register-and-run.sh"]
