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

Prerequisites: 
1. Docker is up and running in your server
2. The configuration file is updated for setting up default parameters for the instance such as organisation details, admin access security etc. This can be changed later as well.

3. Clone this repository to your local server using `git clone`
4. Checkout the latest release or any release available that you wish to run. E.g `git checkout tags/2023.10.4`
5. Execute `make setup`. This sets up the necessary dependencies and configurations for running Consent BB API server instance.
6. Execute `make build`. The compiles and assembles source code into executable files or libraries, following the instructions specified in the Makefile of Consent BB API server instance.
7. Execute `make run`.  This executes a predefined set of instructions in the  Makefile to launch or run the compiled Consent BB API server instance. 

The server is up and running now locally at: "https://api.bb-consent.dev/v2". You can use openAPIs or the admin dashboard to interact with the Consent BB server instance.

## Other resources

* Wiki - https://github.com/decentralised-dataexchange/consent-dev-docs/wiki

## Contributing

Feel free to improve the plugin and send us a pull request. If you find any problems, please create an issue in this repo.

## Licensing
Copyright (c) 2023-25

Licensed under the Apache 2.0 License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the LICENSE for the specific language governing permissions and limitations under the License.
