# Variables
# ---------

OP_VAULT := perf

# Utilies
# ---------

# increment the parameter.
incr = $(shell echo $$(($(1)+1)))

# extract a list of fields secrets values from a 1password item
# usage: $(call op_secrets,my_vault,my_item,my_fields)
op_secrets = $(shell op item get --vault=$(1) --fields=$(shell echo $(3)| sed "s/ /,/g") -- "$(2)")
# extract one secret from secrets returned by op_secrets (space separated)
# usage: $(call op_secret,my_secrets,1)
op_secret = $(shell echo $(1) |cut -d, -f$(2))
# return KEY=VALUE envs list.
op_env = $(eval secrets=$(2)) \
		 $(eval i=1) \
		 $(foreach k,$(1), \
			$(k)=$(shell echo $(secrets)|cut -d',' -f$(i)) \
			$(eval i=$(call incr,($i))))

# Utilies
# ---------

aws_keys = AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY AWS_REGION AWS_BUCKET
aws_fields = username password region bucket
aws_secrets = $(call op_secrets,$(OP_VAULT),AWS S3,$(aws_fields))
aws_env = $(call op_env,$(aws_keys),$(aws_secrets))

# Infos
# -----
.PHONY: help check-op build run

## Display this help screen
help:
	@printf "\e[36m%-35s %s\e[0m\n" "Command" "Usage"
	@sed -n -e '/^## /{'\
		-e 's/## //g;'\
		-e 'h;'\
		-e 'n;'\
		-e 's/:.*//g;'\
		-e 'G;'\
		-e 's/\n/ /g;'\
		-e 'p;}' Makefile | awk '{printf "\033[33m%-35s\033[0m%s\n", $$1, substr($$0,length($$1)+1)}'

check-op:
	@op vault get $(OP_VAULT) > /dev/null

## build program
build:
	@go build

## build and run program locally with env vars from 1password
run: build check-op
	@export $(aws_env) && \
		./perf-fmt --dev $(filter-out $@,$(MAKECMDGOALS))

## echo env vars. usage: $ export $(env)
env: check-op
	@echo $(aws_env)
