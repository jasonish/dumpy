all:
	go build
	rice -v append --exec dumpy || \
		echo "warning: rice not found: static assets won't be bundled"

test:
	@go test $(shell go list ./... | grep -v /vendor/)

install-deps:
	glide install
	npm install moment
	npm install bootstrap
	npm install jquery
	mkdir -p www/vendor
	cp node_modules/bootstrap/dist/css/bootstrap.min.css www/vendor
	cp node_modules/bootstrap/dist/js/bootstrap.min.js www/vendor
	cp node_modules/jquery/dist/jquery.min.js www/vendor
	cp node_modules/jquery/dist/jquery.min.map www/vendor
	cp node_modules/moment/min/moment.min.js www/vendor

dist: VERSION = $(shell ./dumpy version)
dist: GOHOSTARCH = $(shell go env GOHOSTARCH)
dist: GOHOSTOS = $(shell go env GOHOSTOS)
dist: DISTNAME = dumpy-$(VERSION)-$(GOHOSTOS)-$(GOHOSTARCH)
dist: all
	mkdir -p dist/$(DISTNAME)
	cp README.md dist/$(DISTNAME)
	cp LICENSE.txt dist/$(DISTNAME)
	cp dumpy dist/$(DISTNAME)
	cd dist && zip -r $(DISTNAME).zip $(DISTNAME)

clean:
	rm -f dumpy
	rm -rf dist
	find . -name \*~ -exec rm -f {} \;
