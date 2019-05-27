src = src/client.go
bin-path = bin/

build: clean-artefacts
				go fmt ./src/...
				go build -o $(bin-path)aws-keyrotate $(src) 

clean-artefacts: 
				rm -rf $(bin-path)