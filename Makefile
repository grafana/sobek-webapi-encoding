# Run the k6-ci golangci-lint config locally. See grafana/k6-ci/README.md.
# Targets: lint, clean-lint.
# Overrides: WORKFLOW, LINT_BASE, LINT_FINAL.
#
# Produces two gitignored files at the repo root:
#   .golangci-base.yml  cached download from grafana/k6-ci (only re-fetched
#                       when WORKFLOW changes)
#   .golangci.yml       effective config = base (no local patch)

WORKFLOW   ?= .github/workflows/k6-ci.yml
K6_CI_REF  := $(shell grep -oE 'grafana/k6-ci/[^@[:space:]]+@[A-Za-z0-9._/-]+' $(WORKFLOW) | head -n1 | cut -d@ -f2)
BASE_URL   := https://raw.githubusercontent.com/grafana/k6-ci/$(K6_CI_REF)/.golangci.yml

LINT_BASE  ?= .golangci-base.yml
LINT_FINAL ?= .golangci.yml

$(LINT_BASE): $(WORKFLOW)
	curl -fsSL $(BASE_URL) -o $@

$(LINT_FINAL): $(LINT_BASE)
	cp $(LINT_BASE) $@

.PHONY: lint
lint: $(LINT_FINAL)
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$$(head -n1 $(LINT_BASE) | tr -d '# ') \
	  run --config=$(LINT_FINAL) ./...

.PHONY: clean-lint
clean-lint:
	rm -f $(LINT_BASE) $(LINT_FINAL)
