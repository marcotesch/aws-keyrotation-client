src = src/client.go
bin-path = bin/

build: clean-artefacts
				go fmt ./src/...
				go build -o $(bin-path)aws-keyrotate $(src) 

clean-artefacts: 
				rm -rf $(bin-path)

install-mac: build
				cp $(bin-path)aws-keyrotate /usr/local/bin
				cp daemon-examples/com.trc.keyrotation.client.plist $(HOME)/Library/LaunchAgents
				launchctl load $(HOME)/Library/LaunchAgents/com.trc.keyrotation.client.plist