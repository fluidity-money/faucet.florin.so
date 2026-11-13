
PREFIX ?= /usr/local
SYSTEMD_DIR ?= /usr/lib/systemd/system

.PHONY: build cli test run clean

.DELETE_ON_ERROR:

build: cli

CLIS := florin-faucet

UNITS := florin-faucet.socket florin-faucet.service

cli:
	go build -o florin-faucet .

install: cli
	@mkdir -p ${DESTDIR}${PREFIX}/bin
	@install -m755 ${CLIS} ${DESTDIR}${PREFIX}/bin/
	@chmod +x ${DESTDIR}${PREFIX}/bin/${CLIS}
	@install -m644 ${UNITS} ${DESTDIR}${SYSTEMD_DIR}/

clean:
	@rm -f florin-faucet
