#!/usr/bin/env bash

sudo mkdir -p /home/animmendation/tcc_animmendation/dataset_builder/dataset/
sudo mv /tmp/dataset/* /home/animmendation/tcc_animmendation/dataset_builder/dataset/

fish -c "fisher install pure-fish/pure"

exec "$@"