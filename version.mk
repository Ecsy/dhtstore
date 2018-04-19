# ---- program version ----
git-exists: ## Checks git is installed.
	@which git > /dev/null
git-repo: git-exists ## Check project is under git version control.
	@git show > /dev/null

BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
LAST_TAG = $(shell git describe --tags --always)
VER = $(BRANCH)-$(LAST_TAG)

check:: git-repo ## Adds
