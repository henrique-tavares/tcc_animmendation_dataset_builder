#!/usr/bin/env bash

out_dir="$1"
dataset="hernan4444/anime-recommendation-database-2020"
files=("animelist.csv")

for file in ${files[@]}; do
  if [ ! -f "${out_dir}/${file}" ]; then
    kaggle datasets download ${dataset} --unzip -f ${file} -p ${out_dir}/
    unzip ${out_dir}/${file}.zip -d ${out_dir}
    rm ${out_dir}/${file}.zip
  else
    echo "${out_dir}/${file} already exists!"
  fi
done