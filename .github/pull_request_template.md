#### Steps to test this PR

Run the following commands from `bb-consent-playground` repository:
  
1. To destroy all the running containers and their volumes, execute `make destroy`
2. To build fixtures docker image, execute `make build-fixtures`
3. To build BDD test runner docker image, execute `make build-test`
4. In the API repository, switch to the feature/fix branch you wish to execute tests against and build a docker image by executing `make build/docker/deployable`
5. Copy image name and paste it in `image` field value under `api` in `test-docker-compose.yaml` file inside `test` folder in `bb-consent-playground`
