
.PHONY: cache
cache:
	cd api && make build-cache
	cd worker && make build-cache
	cd ship-operator && make build-cache
	cd web && make build-cache

.PHONY: test
test:
	cd web && make test
	cd api && make test

.PHONY: bitbucket-server
bitbucket-server:
	docker volume create --name bitbucketVolume
	@-docker rm -f bitbucket > /dev/null 2>&1 ||:
	docker run \
		-v bitbucketVolume:/var/atlassian/application-data/bitbucket \
		--name="bitbucket" \
		-d \
		-p 7990:7990 \
		-p 7999:7999 \
		atlassian/bitbucket-server:4.12
	@echo "A BitBucket server is starting on http://localhost:7990. You'll need to install an eval license".

