# Note: to make a plugin compatible with a binary built in debug mode, add `-gcflags='all=-N -l'`

PLUGIN_OS ?= linux
PLUGIN_ARCH ?= amd64

plugin_zitadel: bin/$(PLUGIN_OS)$(PLUGIN_ARCH)/zitadel_linux.so

bin/$(PLUGIN_OS)$(PLUGIN_ARCH)/zitadel_linux.so:
	go build -buildmode=plugin -o $@ $^

plugin_zitadel_linux_amd64:
	PLUGIN_OS=linux PLUGIN_ARCH=amd64 make plugin_zitadel

plugin_zitadel_linux_arm64:
	PLUGIN_OS=linux PLUGIN_ARCH=arm64 make plugin_zitadel

plugin_zitadel_darwin_amd64:
	PLUGIN_OS=darwin PLUGIN_ARCH=amd64 make plugin_zitadel

plugin_zitadel_darwin_arm64:
	PLUGIN_OS=darwin PLUGIN_ARCH=arm64 make plugin_zitadel

release-glauth-zitadel: plugin_zitadel_linux_amd64
	mv bin/linuxamd64/zitadel_linux.so bin/zitadel_linux-linux-amd64.so && rmdir bin/linuxamd64

clean:
	rm -rf bin

build_glauth:
	rm -rf gl
	mkdir gl
	cd gl && \
	go build github.com/glauth/glauth/v2

image:
	docker build -t 192.168.6.10:5000/glauth/glauth:latest .
	docker push 192.168.6.10:5000/glauth/glauth:latest
