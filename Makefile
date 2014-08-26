GODEPS :=	github.com/GeertJohan/go.rice \
		github.com/GeertJohan/go.rice/rice \
		code.google.com/p/go.crypto/bcrypt \
		github.com/gorilla/mux \
		gopkg.in/yaml.v1

all:
	go build -o dumpy

dist: VERSION = $(shell ./dumpy version)
dist: GOHOSTARCH = $(shell go env GOHOSTARCH)
dist: GOHOSTOS = $(shell go env GOHOSTOS)
dist: DISTNAME = dumpy-$(VERSION)-$(GOHOSTOS)-$(GOHOSTARCH)
dist: all
	rice -v append --exec dumpy
	mkdir -p dist/$(DISTNAME)
	cp README.md dist/$(DISTNAME)
	cp LICENSE.txt dist/$(DISTNAME)
	cp dumpy dist/$(DISTNAME)
	cd dist && zip -r $(DISTNAME).zip $(DISTNAME)

clean:
	rm -f dumpy
	rm -rf dist
	find . -name \*~ -exec rm -f {} \;

get-go-deps:
	@for dep in $(GODEPS); do \
		echo "go get $$dep"; \
		go get $$dep; \
	done

update-go-deps:
	@for dep in $(GODEPS); do \
		echo "go get -u $$dep"; \
		go get -u $$dep; \
	done

update-bower-components:
	bower update
	cp bower_components/bootstrap/dist/css/bootstrap.min.css www/vendor
	cp bower_components/bootstrap/dist/js/bootstrap.min.js www/vendor
	cp bower_components/jquery/dist/jquery.min.js www/vendor
	cp bower_components/jquery/dist/jquery.min.map www/vendor
	cp bower_components/moment/min/moment.min.js www/vendor
