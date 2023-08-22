# Boostrap development environment


## Usage

 * Add this to your **Makefile**

```
.bootstrap:
    git clone git@github.com:decentralised-dataexchange/bootstrap.git "$(CURDIR)/.bootstrap"

.PHONY: bootstrap
bootstrap: .bootstrap ## Boostraps development environment
    git -C $(CURDIR)/.bootstrap fetch --all --prune
    git -C $(CURDIR)/.bootstrap reset --hard origin/master
    make -C .bootstrap bootstrap
```

 * Add this to your **.gitignore**

```
/.bootstrap/
```
