<h1 align="center">
    GovStack Consent BB API
</h1>

<p align="center">
    <a href="/../../commits/" title="Last Commit"><img src="https://img.shields.io/github/last-commit/decentralised-dataexchange/bb-consent-api?style=flat"></a>
    <a href="/../../issues" title="Open Issues"><img src="https://img.shields.io/github/issues/decentralised-dataexchange/bb-consent-api?style=flat"></a>
    <a href="./LICENSE" title="License"><img src="https://img.shields.io/badge/License-Apache%202.0-yellowgreen?style=flat"></a>
</p>

<p align="center">
  <a href="#about">About</a> •
  <a href="#release-status">Release Status</a> •
  <a href="#contributing">Contributing</a> •
  <a href="#licensing">Licensing</a>
</p>

## About

This repository hosts source code for the reference implementation of GovStack Consent Building Block service APIs

## Release Status

The key deliverables of the project are as given. The table summarises the release status of all the deliverables.

| Identifier | Date          | Deliverable             |
| :--------- | :------------ | :---------------------- |
| D3.1.1     | November 15th | Developer documentation |
| D3.1.2     | November 15th | Test protocol           |

## Instructions to run

The prerequisites for getting the server up and running are as follows:

1. Docker is up and running on your server. You can check it using the command docker ps.
2. Pre-install [jq](https://jqlang.github.io/jq/), a lightweight and flexible command-line JSON processor for parsing and manipulating JSON data.

You can request a pre-defined configuration file and skip the following steps by contacting [support@igrant.io](mailto:support@igrant.io). Please specify the desired admin username in your request. Alternatively, you can proceed with steps 3 and 4 for manual installation and configuration.

3. Install keycloak and use the parameters in step 4 configurations below.
4. The configuration parameters used by the Consent BB API server is created at <server address>/bb-consent-api/resources/config/config-development.json. This sets up the default parameters for the Consent BB server instance, such as organisation details, admin access security (with keycloak, etc. This can also be modified later but would require building the server again.

Note: It is recommended to remove all container instances and volumes running using `docker container rm -f $(docker container ls -aq)` and `docker volume rm $(docker volume ls -q)`. If any of the steps below need to be repeated, we recommend this step to ensure a clean environment.  

Now, follow the steps below to get the ConsentBB API server up and running:  

1. Clone this repository to your local server using `git clone`.
2. Check out the latest release or any available release you wish to run. E.g. `git checkout tags/2023.10.4`.
3. Execute `make setup`. This sets up the necessary dependencies and configurations for running the Consent BB API server instance.
4. Execute `make build`. The compiles and assembles source code into executable files or libraries, following the instructions specified in the Makefile of Consent BB API server instance.
5. Execute `make run`.  This executes a predefined set of instructions in the  Makefile to launch or run the compiled Consent BB API server instance.

The server is up and running now locally at https://api.bb-consent.dev/v2. You can use openAPIs with postman or the admin dashboard to interact with the Consent BB server instance.

## Other resources

* Wiki - https://github.com/decentralised-dataexchange/consent-dev-docs/wiki

## Contributing

Feel free to improve the plugin and send us a pull request. If you find any problems, please create an issue in this repo.

## Licensing
Copyright (c) 2023-25 LCubed AB (iGrant.io), Sweden

Licensed under the Apache 2.0 License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the LICENSE for the specific language governing permissions and limitations under the License.
