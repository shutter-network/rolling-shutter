FROM python:3.10

RUN pip install poetry

COPY tools/snapshot/dummyserver /
WORKDIR /dummyserver

RUN poetry install

ENTRYPOINT ["poetry", "run", "dummyserver"]
