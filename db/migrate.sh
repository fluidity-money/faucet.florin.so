#!/bin/sh -e

table="florin_faucet_migrationst"

dbmate -d migrations --migrations-table "$table" -u "$1" up
