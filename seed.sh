aws dynamodb create-table --endpoint-url http://localhost:8000 --table-name SkranAppTable --attribute-definitions AttributeName=Primary,AttributeType=S AttributeName=Sort,AttributeType=S --key-schema AttributeName=Primary,KeyType=HASH AttributeName=Sort,KeyType=RANGE --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
aws dynamodb batch-write-item \
    --request-items file://data/recipes.json \
    --return-item-collection-metrics SIZE \
    --endpoint-url http://localhost:8000