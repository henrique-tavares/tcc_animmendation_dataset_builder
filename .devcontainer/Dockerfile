FROM python:3.10-slim AS downloader

ARG kaggle_username
ARG kaggle_key

ENV KAGGLE_USERNAME=${kaggle_username}
ENV KAGGLE_KEY=${kaggle_key}

ENV DATASET="hernan4444/anime-recommendation-database-2020"
ENV FILE="animelist.csv"

RUN apt-get update -y && \
  apt-get install -y wget unzip

RUN pip install kaggle

WORKDIR /tmp

COPY download_dataset.sh .

RUN chmod +x ./download_dataset.sh && ./download_dataset.sh ./dataset

FROM golang:1.19

COPY --from=downloader /tmp/dataset /tmp/dataset
