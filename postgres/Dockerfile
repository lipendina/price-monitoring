# пароль прописан явно для удобства проверки задания,
# в публичной базе использовались бы переменные окружения
FROM postgres
ENV POSTGRES_USER docker
ENV POSTGRES_PASSWORD '12345678'
ENV POSTGRES_DB avito
COPY init.sql /docker-entrypoint-initdb.d/