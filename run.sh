# cd ./cmd/web
# go build -o bibService && ./bibService settings.json

cd ./cmd/pod
go build && ./pod ~/data/marc_pod/penn_00705.xml
