FROM docker.elastic.co/elasticsearch/elasticsearch:7.10.2
RUN /usr/share/elasticsearch/bin/elasticsearch-plugin install --batch repository-s3
COPY --chown=elasticsearch:elasticsearch build/elastic/elasticsearch.yml /usr/share/elasticsearch/config/
