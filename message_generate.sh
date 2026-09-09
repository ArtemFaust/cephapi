#!/bin/bash

# $1 - endpoint подключения
# $2 - Квота пользователя
# $3 - Имя пользователя (uuid)
# $4 - access ключ пользователя
# $5 - secret ключ пользователя

# аттрибуты подключения
ENDPOINT_ATTR=$(cat message_templates.json | jq --arg e "$1" '.endpoints[] | select(.[$e]) | .[$e]')
# заготовка письма
TEMPLATE=$(cat message_templates.json | jq '.template' | sed 's/"//g')

# Определение атрибутов подключения
ENDPOINT=$(echo $ENDPOINT_ATTR | jq '.endpoint' | sed 's/"//g')
MAXBUCKETS=$(echo $ENDPOINT_ATTR | jq '.max_buckets' | sed 's/"//g')
REALM=$(echo $ENDPOINT_ATTR | jq '.realm' | sed 's/"//g')
ZONE=$(echo $ENDPOINT_ATTR | jq '.zone' | sed 's/"//g')
ZONE_GROUP=$(echo $ENDPOINT_ATTR | jq '.zone_group' | sed 's/"//g')

# Замена атрибутов подключения
echo -e $TEMPLATE \
    | sed "s|max buckets: |max buckets: $MAXBUCKETS|g" \
    | sed "s|user quota: |user quota: $2|g" \
    | sed "s|username: |username: $3|g" \
    | sed "s|access key: |access key: $4|g" \
    | sed "s|secret key: |secret key: $5|g" \
    | sed "s|realm: |realm: $REALM|g" \
    | sed "s|zone: |zone: $ZONE|g" \
    | sed "s|zone group: |zone group: $ZONE_GROUP|g" \
    | sed "s|endpoints: |endpoints: $ENDPOINT|g"