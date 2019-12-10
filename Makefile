version="0.1.1"
version_file=VERSION
working_dir=$(shell pwd)
arch="armhf"

clean:
	-rm hue-ad

build-go:
	cd ./src;go build -o hue-ad service.go;cd ../

build-go-arm:
	cd ./src;GOOS=linux GOARCH=arm GOARM=6 go build -ldflags="-s -w" -o hue-ad service.go;cd ../

build-go-amd:
	cd ./src;GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o hue-ad service.go;cd ../


configure-arm:
	python ./scripts/config_env.py prod $(version) armhf

configure-amd64:
	python ./scripts/config_env.py prod $(version) amd64


package-tar:
	tar cvzf hue-ad_$(version).tar.gz hue-ad VERSION

package-deb-doc-tp:
	@echo "Packaging application as Thingsplex debian package"
	chmod a+x package/debian_tp/DEBIAN/*
	cp ./src/hue-ad package/debian_tp/opt/thingsplex/hue-ad
	cp VERSION package/debian_tp/opt/thingsplex/hue-ad
	docker run --rm -v ${working_dir}:/build -w /build --name debuild debian dpkg-deb --build package/debian_tp
	@echo "Done"

package-deb-doc-fh:
	@echo "Packaging application as Futurehome debian package"
	chmod a+x package/debian_fh/DEBIAN/*
	cp ./src/hue-ad package/debian_fh/usr/bin/hue-ad
	cp VERSION package/debian_fh/var/lib/futurehome/hue-ad
	docker run --rm -v ${working_dir}:/build -w /build --name debuild debian dpkg-deb --build package/debian_fh
	@echo "Done"


deb-arm-fh : clean configure-arm build-go-arm package-deb-doc-fh
	mv package/debian_fh.deb package/build/hue-ad_$(version)_armhf.deb

deb-arm-tp : clean configure-arm build-go-arm package-deb-doc-tp
	mv package/debian_tp.deb package/build/hue-ad_$(version)_armhf.deb

deb-amd : configure-amd64 build-go-amd package-deb-doc-tp
	mv debian.deb hue-ad_$(version)_amd64.deb

run :
	cd ./src; go run service.go -c testdata/config.json;cd ../


.phony : clean
