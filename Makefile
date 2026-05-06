SHELL := /bin/bash

SPEC ?=
PROFILE ?= smoke
RUN ?=
AGENT ?= codex

.PHONY: refactor-spec refactor-new refactor-validate refactor-prompt agent-run refactor-auto harness report

refactor-spec:
	@if [ -z "$(SPEC)" ]; then echo "Usage: make refactor-spec SPEC=<spec-name>"; exit 1; fi
	@bash scripts/refactor/create_spec.sh "$(SPEC)"

refactor-new:
	@if [ -z "$(SPEC)" ]; then echo "Usage: make refactor-new SPEC=specs/refactor/<name>.md"; exit 1; fi
	@bash scripts/refactor/new_run.sh "$(SPEC)"

refactor-validate:
	@if [ -z "$(SPEC)" ]; then echo "Usage: make refactor-validate SPEC=specs/refactor/<name>.md"; exit 1; fi
	@bash scripts/refactor/validate_spec.sh "$(SPEC)"

refactor-prompt:
	@if [ -z "$(SPEC)" ]; then echo "Usage: make refactor-prompt SPEC=specs/refactor/<name>.md"; exit 1; fi
	@bash scripts/refactor/build_codex_prompt.sh "$(SPEC)"

agent-run:
	@if [ -z "$(RUN)" ]; then echo "Usage: make agent-run RUN=artifacts/refactor_runs/<run-id> [AGENT=codex|claude]"; exit 1; fi
	@bash scripts/refactor/agent_run.sh "$(RUN)" "$(AGENT)" "$(CURDIR)"

refactor-auto:
	@if [ -z "$(SPEC)" ]; then echo "Usage: make refactor-auto SPEC=specs/refactor/<name>.md [AGENT=codex|claude] [PROFILE=smoke|full]"; exit 1; fi
	@bash scripts/refactor/run_workflow.sh "$(SPEC)" "$(AGENT)" "$(PROFILE)"

harness:
	@bash scripts/refactor/harness.sh "$(PROFILE)" "$(RUN)"

report:
	@if [ -z "$(RUN)" ]; then echo "Usage: make report RUN=artifacts/refactor_runs/<run-id>"; exit 1; fi
	@bash scripts/refactor/report.sh "$(RUN)"
