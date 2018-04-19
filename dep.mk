#var GLIDE: ## glide executable path.
DEP ?= dep

dep-exsits: ## Check dep is installed.
	@which dep > /dev/null

Gopkg.toml: ## Creates Gopkg.toml by guessing dependencies from source.
	$(DEP) init -v
	$(MAKE) dep-up

dep-up: dep-exsits ## Update project dependencies (with transitive deps).
	$(DEP) ensure -v
	$(DEP) status

check:: dep-exsits Gopkg.toml
