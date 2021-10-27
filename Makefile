version_file=VERSION
working_dir=$(shell pwd)
arch="armhf"
version:=`git describe --tags | cut -c 2-`
remote_host = "fh@cube.local"

clean:
	-rm ./src/hue-ad

init:
	git config core.hooksPath .githooks

build-go:
	cd ./src;go build -o hue-ad service.go;cd ../

build-go-arm: init
	cd ./src;GOOS=linux GOARCH=arm GOARM=6 go build -ldflags="-s -w" -o hue-ad service.go;cd ../

build-go-amd: init
	cd ./src;GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o hue-ad service.go;cd ../


configure-arm:
	python ./scripts/config_env.py prod $(version) armhf

configure-amd64:
	python ./scripts/config_env.py prod $(version) amd64


package-tar:
	tar cvzf hue-ad_$(version).tar.gz hue-ad $(version_file)

package-deb-doc-tp:
	@echo "Packaging application as Thingsplex debian package"
	chmod a+x package/debian_tp/DEBIAN/*
	mkdir -p package/debian_tp/var/log/thingsplex/hue-ad package/debian_tp/usr/bin
	mkdir -p package/build
	cp ./src/hue-ad package/debian_tp/opt/thingsplex/hue-ad
	cp VERSION package/debian_tp/opt/thingsplex/hue-ad
	docker run --rm -v ${working_dir}:/build -w /build --name debuild debian dpkg-deb --build package/debian_tp
	@echo "Done"


deb-arm : clean configure-arm build-go-arm package-deb-doc-tp
	mv package/debian_tp.deb package/build/hue-ad_$(version)_armhf.deb

deb-amd : configure-amd64 build-go-amd package-deb-doc-tp
	mv debian.deb package/build/hue-ad_$(version)_amd64.deb

upload :
	scp package/build/hue-ad_$(version)_armhf.deb $(remote_host):~/

remote-install : upload
	ssh -t $(remote_host) "sudo dpkg -i hue-ad_$(version)_armhf.deb"

deb-remote-install : deb-arm remote-install
	@echo "Installed"
run :
	cd ./src; go run service.go -c testdata;cd ../

.phony : clean
