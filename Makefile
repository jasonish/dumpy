VIRTUALENV ?=	virtualenv

# Hostname - used for automatic SSL certificate generation.
HOSTNAME ?=	`hostname`

all:

update-jquery:
	curl http://code.jquery.com/jquery-latest.js > \
		dumpy/static/js/jquery-latest.js

setup:
	@if [ ! -e etc/config ]; then \
		echo "Copying etc/config.dist to etc/config."; \
		cp etc/config.dist etc/config; \
	fi

	@if [ ! -e etc/cookie-secret ]; then \
		echo "Creating etc/cookie-secret."; \
		python -c "import os; import hashlib; \
			print(hashlib.sha256(os.urandom(56)).hexdigest())" > \
			etc/cookie-secret; \
	fi

sdist:
	python setup.py sdist

env:
	${VIRTUALENV} ./virtualenv
	. virtualenv/bin/activate; pip install -r etc/pip-freeze.txt

pip-upgrade:
	@. virtualenv/bin/activate; pip install --upgrade -r etc/pip-freeze.txt

pip-freeze:
	@. virtualenv/bin/activate; pip freeze > etc/pip-freeze.txt

test:
	@python ./dumpy/tests/run.py

ssl:
	mkdir -p etc/ssl
	openssl genrsa -des3 \
		-out etc/ssl/server.key.orig \
		-passout pass:password 1024 > /dev/null
	openssl rsa -in etc/ssl/server.key.orig -passin pass:password \
		-out etc/ssl/server.key > /dev/null
	openssl req -new \
		-days 1825 \
	 	-key etc/ssl/server.key \
	 	-out etc/ssl/server.csr \
	 	-subj /CN=${HOSTNAME}
	openssl x509 -req \
		-in etc/ssl/server.csr \
		-signkey etc/ssl/server.key \
		-out etc/ssl/server.crt > /dev/null

clean:
	find . -name \*.pyc | xargs rm -f
	find . -name \*~ | xargs rm -f
	rm -rf MANIFEST dist build
